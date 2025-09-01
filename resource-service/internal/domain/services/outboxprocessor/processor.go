package outboxprocessor

import (
	"context"
	"log/slog"
	"time"

	"github.com/nzb3/diploma/resource-service/internal/domain/models/eventmodel"
)

// eventService defines the interface for event processing operations
type eventService interface {
	GetUnsentEvents(ctx context.Context, limit, offset int) ([]eventmodel.Event, error)
	ProcessEvent(ctx context.Context, event eventmodel.Event) error
}

// Config holds configuration for the outbox processor
type Config struct {
	// Interval specifies how often to check for unsent events
	Interval time.Duration
	// BatchSize specifies the maximum number of events to process in one batch
	BatchSize int
	// MaxRetries specifies the maximum number of retry attempts for failed events
	MaxRetries int
	// RetryDelay specifies the delay between retry attempts
	RetryDelay time.Duration
}

// Processor handles the reliable delivery of events using the outbox pattern
// It periodically scans for unsent events and attempts to publish them
type Processor struct {
	eventService eventService
	config       Config
	stopCh       chan struct{}
	doneCh       chan struct{}
}

// NewOutboxProcessor creates a new outbox processor with the given configuration
func NewOutboxProcessor(eventService eventService, config Config) *Processor {
	if config.Interval == 0 {
		config.Interval = 30 * time.Second
	}
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 5 * time.Second
	}

	return &Processor{
		eventService: eventService,
		config:       config,
		stopCh:       make(chan struct{}),
		doneCh:       make(chan struct{}),
	}
}

// NewDefaultOutboxProcessor creates a new outbox processor with default configuration
func NewDefaultOutboxProcessor(eventService eventService) *Processor {
	return NewOutboxProcessor(eventService, Config{
		Interval:   30 * time.Second,
		BatchSize:  100,
		MaxRetries: 3,
		RetryDelay: 5 * time.Second,
	})
}

// Start begins the outbox processor background operation
// This method blocks until Stop is called or the context is cancelled
func (p *Processor) Start(ctx context.Context) {
	defer close(p.doneCh)

	ticker := time.NewTicker(p.config.Interval)
	defer ticker.Stop()

	slog.InfoContext(ctx, "Starting outbox processor",
		"interval", p.config.Interval,
		"batch_size", p.config.BatchSize,
		"max_retries", p.config.MaxRetries,
		"retry_delay", p.config.RetryDelay)

	for {
		select {
		case <-ctx.Done():
			slog.InfoContext(ctx, "Outbox processor stopped due to context cancellation")
			return
		case <-p.stopCh:
			slog.InfoContext(ctx, "Outbox processor stopped")
			return
		case <-ticker.C:
			p.processEvents(ctx)
		}
	}
}

// Stop gracefully stops the outbox processor
func (p *Processor) Stop() {
	close(p.stopCh)
	<-p.doneCh
}

// processEvents processes a batch of unsent events
func (p *Processor) processEvents(ctx context.Context) {
	const op = "OutboxProcessor.processEvents"

	events, err := p.eventService.GetUnsentEvents(ctx, p.config.BatchSize, 0)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get unsent events",
			"op", op,
			"error", err)
		return
	}

	if len(events) == 0 {
		return
	}

	slog.InfoContext(ctx, "Processing unsent events",
		"op", op,
		"count", len(events))

	successCount := 0
	failureCount := 0

	for _, event := range events {
		err := p.processEventWithRetry(ctx, event)
		if err != nil {
			failureCount++
			slog.ErrorContext(ctx, "Failed to process event after retries",
				"op", op,
				"error", err,
				"event_id", event.ID,
				"event_name", event.Name)
		} else {
			successCount++
		}
	}

	slog.InfoContext(ctx, "Batch processing completed",
		"op", op,
		"total", len(events),
		"success", successCount,
		"failed", failureCount)
}

// processEventWithRetry attempts to process an event with retry logic
func (p *Processor) processEventWithRetry(ctx context.Context, event eventmodel.Event) error {
	const op = "OutboxProcessor.processEventWithRetry"

	var lastErr error

	for attempt := 1; attempt <= p.config.MaxRetries; attempt++ {
		err := p.eventService.ProcessEvent(ctx, event)
		if err == nil {
			if attempt > 1 {
				slog.InfoContext(ctx, "Event processed successfully after retries",
					"op", op,
					"event_id", event.ID,
					"attempt", attempt)
			}
			return nil
		}

		lastErr = err
		slog.WarnContext(ctx, "Failed to process event, will retry",
			"op", op,
			"error", err,
			"event_id", event.ID,
			"event_name", event.Name,
			"attempt", attempt,
			"max_retries", p.config.MaxRetries)

		if attempt < p.config.MaxRetries {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(p.config.RetryDelay):
			}
		}
	}

	return lastErr
}

// ProcessNow immediately processes any pending events (useful for testing or manual triggers)
func (p *Processor) ProcessNow(ctx context.Context) error {
	const op = "OutboxProcessor.ProcessNow"

	slog.InfoContext(ctx, "Manual processing of unsent events triggered", "op", op)
	p.processEvents(ctx)
	return nil
}
