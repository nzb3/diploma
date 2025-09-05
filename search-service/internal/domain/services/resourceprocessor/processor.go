package resourceprocessor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"

	"github.com/nzb3/diploma/search-service/internal/domain/models"
	"github.com/nzb3/diploma/search-service/internal/repository/messaging"
)

// vectorStorage defines the interface for vector storage operations
type vectorStorage interface {
	PutResource(ctx context.Context, resource models.Resource) ([]string, error)
}

// eventService defines the interface for event publishing operations
type eventService interface {
	PublishEvent(ctx context.Context, topic string, eventName string, data interface{}) error
}

// IndexationCompleteEvent represents the event published after indexation
type IndexationCompleteEvent struct {
	ResourceID uuid.UUID `json:"resource_id"`
	Success    bool      `json:"success"`
	Message    string    `json:"message"`
	ChunkIDs   []string  `json:"chunk_ids,omitempty"`
}

// Processor handles resource indexation events from the resource-service
type Processor struct {
	vectorStorage vectorStorage
	eventService  eventService
	consumer      messaging.MessageConsumer
	stopCh        chan struct{}
	doneCh        chan struct{}
	wg            sync.WaitGroup
}

// NewResourceProcessor creates a new resource processor
func NewResourceProcessor(
	vectorStorage vectorStorage,
	eventService eventService,
	consumer messaging.MessageConsumer,
) *Processor {
	return &Processor{
		vectorStorage: vectorStorage,
		eventService:  eventService,
		consumer:      consumer,
		stopCh:        make(chan struct{}),
		doneCh:        make(chan struct{}),
	}
}

// Start begins listening for resource created events
func (p *Processor) Start(ctx context.Context) error {
	defer close(p.doneCh)

	topics := []string{"resource"}

	err := p.consumer.Subscribe(ctx, topics, p)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topics: %w", err)
	}

	slog.InfoContext(ctx, "Resource processor started", "topics", topics)

	select {
	case <-ctx.Done():
		slog.InfoContext(ctx, "Resource processor stopped due to context cancellation")
	case <-p.stopCh:
		slog.InfoContext(ctx, "Resource processor stopped")
	}

	return nil
}

// Stop gracefully stops the resource processor
func (p *Processor) Stop() {
	close(p.stopCh)

	if p.consumer != nil {
		p.consumer.Close()
	}

	p.wg.Wait()
	<-p.doneCh
}

// HandleMessage implements the MessageHandler interface
func (p *Processor) HandleMessage(ctx context.Context, topic string, key string, value []byte, headers map[string]string) error {
	const op = "ResourceProcessor.HandleMessage"

	if topic != "resource" {
		return nil
	}

	p.wg.Add(1)
	defer p.wg.Done()

	slog.DebugContext(ctx, "Processing resource message",
		"topic", topic,
		"key", key,
		"headers", headers)

	// Check if this is a resource.created event
	eventName, exists := headers["event-name"]
	if !exists || eventName != "resource.created" {
		slog.DebugContext(ctx, "Ignoring non-resource.created event",
			"event_name", eventName)
		return nil
	}

	// Parse the resource from the message payload
	var resource models.Resource
	if err := json.Unmarshal(value, &resource); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal resource",
			"op", op,
			"error", err)
		return fmt.Errorf("%s: failed to unmarshal resource: %w", op, err)
	}

	slog.InfoContext(ctx, "Processing resource for indexation",
		"resource_id", resource.ID,
		"resource_name", resource.Name,
		"resource_type", resource.Type)

	// Process the resource
	chunkIDs, err := p.processResource(ctx, resource)
	if err != nil {
		// Publish failure event
		p.publishIndexationEvent(ctx, resource.ID, false, err.Error(), nil)
		return fmt.Errorf("%s: failed to process resource: %w", op, err)
	}

	// Publish success event
	p.publishIndexationEvent(ctx, resource.ID, true, "Resource indexed successfully", chunkIDs)

	slog.InfoContext(ctx, "Resource processed successfully",
		"resource_id", resource.ID,
		"chunks_count", len(chunkIDs))

	return nil
}

// processResource handles the actual resource processing
func (p *Processor) processResource(ctx context.Context, resource models.Resource) ([]string, error) {
	const op = "ResourceProcessor.processResource"

	slog.DebugContext(ctx, "Starting resource processing",
		"resource_id", resource.ID,
		"content_length", len(resource.ExtractedContent))

	// Use the PutResource method to store the resource in vector storage
	chunkIDs, err := p.vectorStorage.PutResource(ctx, resource)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to store resource in vector storage",
			"op", op,
			"resource_id", resource.ID,
			"error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	slog.InfoContext(ctx, "Resource stored in vector storage",
		"resource_id", resource.ID,
		"chunks_created", len(chunkIDs))

	return chunkIDs, nil
}

// publishIndexationEvent publishes the indexation complete event
func (p *Processor) publishIndexationEvent(ctx context.Context, resourceID uuid.UUID, success bool, message string, chunkIDs []string) {
	const op = "ResourceProcessor.publishIndexationEvent"

	event := IndexationCompleteEvent{
		ResourceID: resourceID,
		Success:    success,
		Message:    message,
		ChunkIDs:   chunkIDs,
	}

	err := p.eventService.PublishEvent(ctx, "indexation_complete", "indexation_complete", event)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to publish indexation complete event",
			"op", op,
			"resource_id", resourceID,
			"success", success,
			"error", err)
		// Don't return error here as the resource processing might have succeeded
		return
	}

	slog.InfoContext(ctx, "Indexation complete event published",
		"resource_id", resourceID,
		"success", success)
}

// Health checks the health of the resource processor
func (p *Processor) Health(ctx context.Context) error {
	return p.consumer.Health(ctx)
}
