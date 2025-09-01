package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/IBM/sarama"

	"github.com/nzb3/diploma/resource-service/internal/repository/messaging"
)

// Consumer implements the MessageConsumer interface using Apache Kafka
type Consumer struct {
	consumer sarama.ConsumerGroup
	config   *ConsumerConfig
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// ConsumerConfig holds the configuration for Kafka consumer
type ConsumerConfig struct {
	Brokers         []string
	GroupID         string
	AutoOffsetReset string // earliest, latest
}

// NewKafkaConsumer creates a new Kafka consumer with the given configuration
func NewKafkaConsumer(config *ConsumerConfig) (*Consumer, error) {
	if config == nil {
		return nil, fmt.Errorf("kafka consumer config cannot be nil")
	}

	if len(config.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers list cannot be empty")
	}

	if config.GroupID == "" {
		return nil, fmt.Errorf("kafka consumer group ID cannot be empty")
	}

	// Create Sarama configuration
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaConfig.Consumer.Return.Errors = true

	// Set auto offset reset
	if config.AutoOffsetReset == "earliest" {
		saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	} else {
		saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest
	}

	// Create the consumer group
	consumer, err := sarama.NewConsumerGroup(config.Brokers, config.GroupID, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer group: %w", err)
	}

	return &Consumer{
		consumer: consumer,
		config:   config,
	}, nil
}

// Subscribe subscribes to topics and starts consuming messages
func (c *Consumer) Subscribe(ctx context.Context, topics []string, handler messaging.MessageHandler) error {
	if len(topics) == 0 {
		return fmt.Errorf("topics list cannot be empty")
	}

	if handler == nil {
		return fmt.Errorf("message handler cannot be nil")
	}

	// Create context with cancellation
	consumerCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	// Create consumer group handler
	groupHandler := &consumerGroupHandler{
		handler: handler,
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		defer cancel()

		for {
			select {
			case <-consumerCtx.Done():
				slog.Info("Kafka consumer context cancelled")
				return
			default:
				// Consume should be called inside an infinite loop
				if err := c.consumer.Consume(consumerCtx, topics, groupHandler); err != nil {
					slog.Error("Error from consumer", "error", err)
					return
				}
			}
		}
	}()

	// Handle consumer errors
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-consumerCtx.Done():
				return
			case err := <-c.consumer.Errors():
				if err != nil {
					slog.Error("Consumer error", "error", err)
				}
			}
		}
	}()

	slog.Info("Kafka consumer started",
		"topics", topics,
		"group_id", c.config.GroupID)

	return nil
}

// Health checks if the consumer can communicate with Kafka brokers
func (c *Consumer) Health(ctx context.Context) error {
	// Create a simple health check by trying to get metadata
	client, err := sarama.NewClient(c.config.Brokers, sarama.NewConfig())
	if err != nil {
		return fmt.Errorf("failed to create kafka client for health check: %w", err)
	}
	defer client.Close()

	// Check if we can get broker information
	brokers := client.Brokers()
	if len(brokers) == 0 {
		return fmt.Errorf("no kafka brokers available")
	}

	// Try to connect to at least one broker
	for _, broker := range brokers {
		connected, err := broker.Connected()
		if err == nil && connected {
			return nil
		}
	}

	return fmt.Errorf("cannot connect to any kafka broker")
}

// Close gracefully shuts down the consumer
func (c *Consumer) Close() error {
	if c.cancel != nil {
		c.cancel()
	}

	// Wait for all goroutines to finish
	c.wg.Wait()

	if c.consumer != nil {
		return c.consumer.Close()
	}
	return nil
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler
type consumerGroupHandler struct {
	handler messaging.MessageHandler
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// Handle messages
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			// Convert headers to map
			headers := make(map[string]string)
			for _, header := range message.Headers {
				headers[string(header.Key)] = string(header.Value)
			}

			// Handle the message
			err := h.handler.HandleMessage(
				session.Context(),
				message.Topic,
				string(message.Key),
				message.Value,
				headers,
			)

			if err != nil {
				slog.Error("Error handling message",
					"topic", message.Topic,
					"partition", message.Partition,
					"offset", message.Offset,
					"error", err)
				// Depending on your error handling strategy, you might want to:
				// 1. Continue processing (current behavior)
				// 2. Return error to stop processing this partition
				// 3. Mark message and continue
			}

			// Mark message as processed
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}

// NewDefaultConsumerConfig returns a default Kafka consumer configuration
func NewDefaultConsumerConfig(brokers []string, groupID string) *ConsumerConfig {
	return &ConsumerConfig{
		Brokers:         brokers,
		GroupID:         groupID,
		AutoOffsetReset: "latest",
	}
}
