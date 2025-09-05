package messaging

import (
	"context"

	"github.com/nzb3/diploma/search-service/internal/domain/models/eventmodel"
)

// MessageHandler defines the interface for handling consumed messages
type MessageHandler interface {
	// HandleMessage processes a consumed message
	HandleMessage(ctx context.Context, topic string, key string, value []byte, headers map[string]string) error
}

// MessageProducer defines the interface for publishing messages to a message broker
// This interface abstracts the underlying messaging implementation (Kafka, RabbitMQ, etc.)
// to ensure loose coupling, testability, and future flexibility
type MessageProducer interface {
	// PublishEvent publishes an event to the configured message broker
	// Returns an error if the publishing fails
	PublishEvent(ctx context.Context, event eventmodel.Event) error

	// Close gracefully shuts down the producer and releases resources
	Close() error

	// Health checks if the producer is healthy and can communicate with the broker
	Health(ctx context.Context) error
}

// MessageConsumer defines the interface for consuming messages from a message broker
// This interface abstracts the underlying messaging implementation (Kafka, RabbitMQ, etc.)
// to ensure loose coupling, testability, and future flexibility
type MessageConsumer interface {
	// Subscribe subscribes to one or more topics and starts consuming messages
	// The handler will be called for each consumed message
	Subscribe(ctx context.Context, topics []string, handler MessageHandler) error

	// Close gracefully shuts down the consumer and releases resources
	Close() error

	// Health checks if the consumer is healthy and can communicate with the broker
	Health(ctx context.Context) error
}
