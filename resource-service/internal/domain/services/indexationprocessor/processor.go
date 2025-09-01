package indexationprocessor

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/google/uuid"

	"github.com/nzb3/diploma/resource-service/internal/domain/models/resourcemodel"
	"github.com/nzb3/diploma/resource-service/internal/repository/messaging"
)

// IndexationCompleteEvent represents the event payload for indexation completion
type IndexationCompleteEvent struct {
	ResourceID uuid.UUID `json:"resource_id"`
	Success    bool      `json:"success"`
	Message    string    `json:"message,omitempty"`
}

// resourceService defines the interface for updating resource status and managing channels
type resourceService interface {
	UpdateResourceStatus(ctx context.Context, resource resourcemodel.Resource, status resourcemodel.ResourceStatus) (resourcemodel.Resource, error)
	GetResourceStatusChannel(resourceID uuid.UUID) (chan resourcemodel.ResourceStatusUpdate, bool)
	RemoveResourceStatusChannel(resourceID uuid.UUID)
	GetResourceByID(ctx context.Context, resourceID uuid.UUID) (resourcemodel.Resource, error)
}

// Processor handles indexation completion events and updates resource status
type Processor struct {
	resourceService resourceService
	consumer        messaging.MessageConsumer
	stopCh          chan struct{}
	doneCh          chan struct{}
	wg              sync.WaitGroup
}

// NewIndexationProcessor creates a new indexation completion processor
func NewIndexationProcessor(resourceService resourceService, consumer messaging.MessageConsumer) *Processor {
	return &Processor{
		resourceService: resourceService,
		consumer:        consumer,
		stopCh:          make(chan struct{}),
		doneCh:          make(chan struct{}),
	}
}

// Start begins listening for indexation completion events
func (p *Processor) Start(ctx context.Context) error {
	defer close(p.doneCh)

	topics := []string{"indexation_complete"}

	err := p.consumer.Subscribe(ctx, topics, p)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topics: %w", err)
	}

	slog.InfoContext(ctx, "Indexation processor started", "topics", topics)

	select {
	case <-ctx.Done():
		slog.InfoContext(ctx, "Indexation processor stopped due to context cancellation")
	case <-p.stopCh:
		slog.InfoContext(ctx, "Indexation processor stopped")
	}

	return nil
}

// Stop gracefully stops the indexation processor
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
	const op = "IndexationProcessor.HandleMessage"

	if topic != "indexation_complete" {
		return nil
	}

	p.wg.Add(1)
	defer p.wg.Done()

	slog.DebugContext(ctx, "Processing indexation complete event",
		"op", op,
		"topic", topic,
		"key", key)

	var event IndexationCompleteEvent
	if err := json.Unmarshal(value, &event); err != nil {
		slog.ErrorContext(ctx, "Failed to unmarshal indexation complete event",
			"op", op,
			"error", err,
			"payload", string(value))
		return fmt.Errorf("%s: failed to unmarshal event: %w", op, err)
	}

	slog.InfoContext(ctx, "Received indexation complete event",
		"op", op,
		"resource_id", event.ResourceID,
		"success", event.Success,
		"message", event.Message)

	resource, err := p.resourceService.GetResourceByID(ctx, event.ResourceID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get resource for status update",
			"op", op,
			"resource_id", event.ResourceID,
			"error", err)
		return fmt.Errorf("%s: failed to get resource: %w", op, err)
	}

	var finalStatus resourcemodel.ResourceStatus
	if event.Success {
		finalStatus = resourcemodel.ResourceStatusCompleted
	} else {
		finalStatus = resourcemodel.ResourceStatusFailed
	}

	_, err = p.resourceService.UpdateResourceStatus(ctx, resource, finalStatus)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update resource status",
			"op", op,
			"resource_id", event.ResourceID,
			"status", finalStatus,
			"error", err)
		return fmt.Errorf("%s: failed to update resource status: %w", op, err)
	}

	slog.InfoContext(ctx, "Updated resource status",
		"op", op,
		"resource_id", event.ResourceID,
		"old_status", resource.Status,
		"new_status", finalStatus)

	statusCh, exists := p.resourceService.GetResourceStatusChannel(event.ResourceID)
	if exists {
		statusUpdate := resourcemodel.ResourceStatusUpdate{
			ResourceID: event.ResourceID,
			Status:     finalStatus,
		}

		select {
		case statusCh <- statusUpdate:
			slog.InfoContext(ctx, "Sent status update to channel",
				"op", op,
				"resource_id", event.ResourceID,
				"status", finalStatus)
		case <-ctx.Done():
			slog.WarnContext(ctx, "Context cancelled while sending status update",
				"op", op,
				"resource_id", event.ResourceID)
			return ctx.Err()
		default:
			slog.WarnContext(ctx, "Status channel is full, dropping update",
				"op", op,
				"resource_id", event.ResourceID)
		}

		close(statusCh)
		p.resourceService.RemoveResourceStatusChannel(event.ResourceID)

		slog.InfoContext(ctx, "Closed and removed status channel",
			"op", op,
			"resource_id", event.ResourceID)
	} else {
		slog.WarnContext(ctx, "No status channel found for resource",
			"op", op,
			"resource_id", event.ResourceID)
	}

	slog.InfoContext(ctx, "Successfully processed indexation complete event",
		"op", op,
		"resource_id", event.ResourceID,
		"final_status", finalStatus)

	return nil
}

// Health checks the health of the indexation processor
func (p *Processor) Health(ctx context.Context) error {
	if p.consumer != nil {
		return p.consumer.Health(ctx)
	}
	return fmt.Errorf("consumer not initialized")
}
