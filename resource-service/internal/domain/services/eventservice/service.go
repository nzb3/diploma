package eventservice

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/nzb3/diploma/resource-service/internal/domain/models/eventmodel"
)

// eventRepository defines the interface for event persistence operations
type eventRepository interface {
	CreateEvent(ctx context.Context, event eventmodel.Event) (eventmodel.Event, error)
	GetNotSentEvents(ctx context.Context, limit int, offset int) ([]eventmodel.Event, error)
	MarkEventAsSent(ctx context.Context, eventID uuid.UUID) error
}

// messageProducer defines the interface for publishing messages
type messageProducer interface {
	PublishEvent(ctx context.Context, event eventmodel.Event) error
	Health(ctx context.Context) error
}

// Service implements the event service using the outbox pattern
// This ensures reliable event publishing by first storing events in the database
// and then publishing them to the message broker
type Service struct {
	eventRepo eventRepository
	producer  messageProducer
}

// NewEventService creates a new event service instance
func NewEventService(repo eventRepository, producer messageProducer) *Service {
	return &Service{
		eventRepo: repo,
		producer:  producer,
	}
}

// PublishEvent publishes a resource-related event using the outbox pattern
// This method ensures ACID properties by storing the event in the same transaction
// as the business operation and then attempting immediate delivery
func (s *Service) PublishEvent(ctx context.Context, topic string, eventName string, data interface{}) error {
	const op = "EventService.PublishEvent"

	event, err := eventmodel.NewEvent(eventName, topic, data)
	if err != nil {
		return fmt.Errorf("%s: failed to create event: %w", op, err)
	}

	savedEvent, err := s.eventRepo.CreateEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("%s: failed to save event to outbox: %w", op, err)
	}

	err = s.producer.PublishEvent(ctx, savedEvent)
	if err != nil {
		slog.WarnContext(ctx, "Failed to publish event immediately, will retry via outbox processor",
			"error", err,
			"event_id", savedEvent.ID,
			"event_name", savedEvent.Name)
		return nil
	}

	err = s.eventRepo.MarkEventAsSent(ctx, savedEvent.ID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to mark event as sent after successful publish",
			"error", err,
			"event_id", savedEvent.ID,
			"event_name", savedEvent.Name)
	}

	slog.InfoContext(ctx, "Event published successfully",
		"event_id", savedEvent.ID,
		"event_name", savedEvent.Name,
		"topic", savedEvent.Topic)

	return nil
}

// GetUnsentEvents retrieves events that haven't been successfully published
// This is used by the outbox processor for retry logic
func (s *Service) GetUnsentEvents(ctx context.Context, limit, offset int) ([]eventmodel.Event, error) {
	const op = "EventService.GetUnsentEvents"

	events, err := s.eventRepo.GetNotSentEvents(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to retrieve unsent events: %w", op, err)
	}

	return events, nil
}

// ProcessEvent attempts to publish a single event and marks it as sent if successful
// This is used by the outbox processor
func (s *Service) ProcessEvent(ctx context.Context, event eventmodel.Event) error {
	const op = "EventService.ProcessEvent"

	err := s.producer.PublishEvent(ctx, event)
	if err != nil {
		return fmt.Errorf("%s: failed to publish event: %w", op, err)
	}

	err = s.eventRepo.MarkEventAsSent(ctx, event.ID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to mark event as sent after successful publish",
			"error", err,
			"event_id", event.ID,
			"event_name", event.Name)
	}

	return nil
}

// Health checks the health of the event service dependencies
func (s *Service) Health(ctx context.Context) error {
	if err := s.producer.Health(ctx); err != nil {
		return fmt.Errorf("message producer health check failed: %w", err)
	}

	return nil
}
