package kafka

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"

	"github.com/nzb3/diploma/search-service/internal/domain/models/eventmodel"
)

// Producer implements the MessageProducer interface using Apache Kafka
type Producer struct {
	producer sarama.SyncProducer
	config   *Config
}

// Config holds the configuration for Kafka producer
type Config struct {
	Brokers []string
	// Additional configuration options can be added here
	RequiredAcks    sarama.RequiredAcks
	RetryMax        int
	CompressionType sarama.CompressionCodec
}

// NewKafkaProducer creates a new Kafka producer with the given configuration
func NewKafkaProducer(config *Config) (*Producer, error) {
	if config == nil {
		return nil, fmt.Errorf("kafka config cannot be nil")
	}

	if len(config.Brokers) == 0 {
		return nil, fmt.Errorf("kafka brokers list cannot be empty")
	}

	// Create Sarama configuration
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.Producer.Retry.Max = config.RetryMax
	saramaConfig.Producer.RequiredAcks = config.RequiredAcks
	saramaConfig.Producer.Compression = config.CompressionType

	saramaConfig.Net.MaxOpenRequests = 1

	// Create the producer
	producer, err := sarama.NewSyncProducer(config.Brokers, saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return &Producer{
		producer: producer,
		config:   config,
	}, nil
}

// PublishEvent publishes an event to Kafka
func (p *Producer) PublishEvent(ctx context.Context, event eventmodel.Event) error {
	// Create Kafka message
	message := &sarama.ProducerMessage{
		Topic: event.Topic,
		Key:   sarama.StringEncoder(event.ID.String()),
		Value: sarama.ByteEncoder(event.Payload),
		Headers: []sarama.RecordHeader{
			{Key: []byte("event_name"), Value: []byte(event.Name)},
			{Key: []byte("event_id"), Value: []byte(event.ID.String())},
			{Key: []byte("event_time"), Value: []byte(event.EventTime.Format("2006-01-02T15:04:05Z"))},
		},
	}

	// Send message
	partition, offset, err := p.producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed to publish event to kafka: %w", err)
	}

	slog.InfoContext(ctx, "Event published to Kafka successfully",
		"topic", event.Topic,
		"partition", partition,
		"offset", offset,
		"event_id", event.ID,
		"event_name", event.Name)

	return nil
}

// Health checks if the producer can communicate with Kafka brokers
func (p *Producer) Health(ctx context.Context) error {
	// Create a simple health check by trying to get metadata
	client, err := sarama.NewClient(p.config.Brokers, sarama.NewConfig())
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

// Close gracefully shuts down the producer
func (p *Producer) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}

// NewDefaultConfig returns a default Kafka producer configuration
func NewDefaultConfig(brokers []string) *Config {
	return &Config{
		Brokers:         brokers,
		RequiredAcks:    sarama.WaitForAll, // Wait for all replicas
		RetryMax:        3,
		CompressionType: sarama.CompressionSnappy,
	}
}
