package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/IBM/sarama"

	"github.com/nzb3/diploma/search-service/internal/repository/messaging"
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

// NewDefaultConsumerConfig returns a consumer configuration with sensible defaults
func NewDefaultConsumerConfig(brokers []string, groupID string) *ConsumerConfig {
	return &ConsumerConfig{
		Brokers:         brokers,
		GroupID:         groupID,
		AutoOffsetReset: "earliest",
	}
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

	// Start error handling goroutine
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			select {
			case <-consumerCtx.Done():
				return
			case err := <-c.consumer.Errors():
				if err != nil {
					slog.Error("Kafka consumer error", "error", err)
				}
			}
		}
	}()

	slog.Info("Kafka consumer subscribed to topics", "topics", topics, "group_id", c.config.GroupID)
	return nil
}

// Health checks if the consumer is healthy
func (c *Consumer) Health(ctx context.Context) error {
	// For Kafka consumer, we can check if the consumer group is still active
	// In a real implementation, you might want to check broker connectivity
	if c.consumer == nil {
		return fmt.Errorf("kafka consumer is not initialized")
	}
	return nil
}

// Close gracefully shuts down the consumer
func (c *Consumer) Close() error {
	slog.Info("Closing Kafka consumer")

	// Cancel the context to stop consuming
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
					"key", string(message.Key),
					"error", err)
				// Don't return error to continue processing other messages
			} else {
				// Mark message as processed only if no error
				session.MarkMessage(message, "")
			}

		case <-session.Context().Done():
			return nil
		}
	}
}
