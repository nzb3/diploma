package outboxprocessor

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/nzb3/diploma/resource-service/internal/domain/models/eventmodel"
)

// MockEventService is a simple mock implementation of the eventService interface
type MockEventService struct {
	mu                       sync.Mutex
	getUnsentEventsResponse  []eventmodel.Event
	getUnsentEventsError     error
	processEventError        error
	getUnsentEventsCalls     int
	processEventCalls        int
	processedEvents          []eventmodel.Event
	processEventErrorMap     map[string]error // Map event ID to error for more control
	processEventCallSequence []error          // Sequence of errors to return on successive calls
	processEventCallIndex    int
}

func (m *MockEventService) GetUnsentEvents(ctx context.Context, limit, offset int) ([]eventmodel.Event, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getUnsentEventsCalls++
	return m.getUnsentEventsResponse, m.getUnsentEventsError
}

func (m *MockEventService) ProcessEvent(ctx context.Context, event eventmodel.Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processEventCalls++
	m.processedEvents = append(m.processedEvents, event)

	// Check for specific event errors
	if m.processEventErrorMap != nil {
		if err, exists := m.processEventErrorMap[event.ID.String()]; exists {
			return err
		}
	}

	// Check for sequence errors
	if m.processEventCallSequence != nil && m.processEventCallIndex < len(m.processEventCallSequence) {
		err := m.processEventCallSequence[m.processEventCallIndex]
		m.processEventCallIndex++
		return err
	}

	return m.processEventError
}

func (m *MockEventService) GetProcessEventCallCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.processEventCalls
}

func (m *MockEventService) GetProcessedEvents() []eventmodel.Event {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]eventmodel.Event{}, m.processedEvents...)
}

func (m *MockEventService) SetProcessEventErrorForEvent(eventID string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.processEventErrorMap == nil {
		m.processEventErrorMap = make(map[string]error)
	}
	m.processEventErrorMap[eventID] = err
}

func (m *MockEventService) SetProcessEventCallSequence(errors []error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processEventCallSequence = errors
	m.processEventCallIndex = 0
}

func TestNewOutboxProcessor(t *testing.T) {
	mockService := &MockEventService{}

	t.Run("with custom config", func(t *testing.T) {
		config := Config{
			Interval:   10 * time.Second,
			BatchSize:  50,
			MaxRetries: 5,
			RetryDelay: 2 * time.Second,
		}

		processor := NewOutboxProcessor(mockService, config)

		if processor == nil {
			t.Error("processor should not be nil")
		}
		if processor.eventService != mockService {
			t.Error("eventService should match mockService")
		}
		if processor.config.Interval != config.Interval {
			t.Errorf("expected interval %v, got %v", config.Interval, processor.config.Interval)
		}
		if processor.config.BatchSize != config.BatchSize {
			t.Errorf("expected batch size %d, got %d", config.BatchSize, processor.config.BatchSize)
		}
		if processor.config.MaxRetries != config.MaxRetries {
			t.Errorf("expected max retries %d, got %d", config.MaxRetries, processor.config.MaxRetries)
		}
		if processor.config.RetryDelay != config.RetryDelay {
			t.Errorf("expected retry delay %v, got %v", config.RetryDelay, processor.config.RetryDelay)
		}
	})

	t.Run("with zero values should use defaults", func(t *testing.T) {
		config := Config{}

		processor := NewOutboxProcessor(mockService, config)

		if processor.config.Interval != 30*time.Second {
			t.Errorf("expected default interval 30s, got %v", processor.config.Interval)
		}
		if processor.config.BatchSize != 100 {
			t.Errorf("expected default batch size 100, got %d", processor.config.BatchSize)
		}
		if processor.config.MaxRetries != 3 {
			t.Errorf("expected default max retries 3, got %d", processor.config.MaxRetries)
		}
		if processor.config.RetryDelay != 5*time.Second {
			t.Errorf("expected default retry delay 5s, got %v", processor.config.RetryDelay)
		}
	})
}

func TestNewDefaultOutboxProcessor(t *testing.T) {
	mockService := &MockEventService{}

	processor := NewDefaultOutboxProcessor(mockService)

	if processor == nil {
		t.Error("processor should not be nil")
	}
	if processor.eventService != mockService {
		t.Error("eventService should match mockService")
	}
	if processor.config.Interval != 30*time.Second {
		t.Errorf("expected default interval 30s, got %v", processor.config.Interval)
	}
	if processor.config.BatchSize != 100 {
		t.Errorf("expected default batch size 100, got %d", processor.config.BatchSize)
	}
	if processor.config.MaxRetries != 3 {
		t.Errorf("expected default max retries 3, got %d", processor.config.MaxRetries)
	}
	if processor.config.RetryDelay != 5*time.Second {
		t.Errorf("expected default retry delay 5s, got %v", processor.config.RetryDelay)
	}
}

func TestProcessor_Start_ContextCancellation(t *testing.T) {
	mockService := &MockEventService{
		getUnsentEventsResponse: []eventmodel.Event{},
		getUnsentEventsError:    nil,
	}

	config := Config{
		Interval:   50 * time.Millisecond,
		BatchSize:  100,
		MaxRetries: 3,
		RetryDelay: 1 * time.Millisecond,
	}

	processor := NewOutboxProcessor(mockService, config)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		processor.Start(ctx)
		close(done)
	}()

	select {
	case <-done:
		// Processor should stop when context is cancelled
	case <-time.After(300 * time.Millisecond):
		t.Fatal("Processor didn't stop after context cancellation")
	}

	if mockService.getUnsentEventsCalls == 0 {
		t.Error("GetUnsentEvents should have been called at least once")
	}
}

func TestProcessor_Start_StopGracefully(t *testing.T) {
	mockService := &MockEventService{
		getUnsentEventsResponse: []eventmodel.Event{},
		getUnsentEventsError:    nil,
	}

	config := Config{
		Interval:   50 * time.Millisecond,
		BatchSize:  100,
		MaxRetries: 3,
		RetryDelay: 1 * time.Millisecond,
	}

	processor := NewOutboxProcessor(mockService, config)

	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processor.Start(ctx)
		close(done)
	}()

	// Wait a bit to let processor start
	time.Sleep(25 * time.Millisecond)

	// Stop the processor
	processor.Stop()

	// Wait for processor to stop
	select {
	case <-done:
		// Processor should stop
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Processor didn't stop gracefully")
	}
}

func TestProcessor_processEvents_NoEvents(t *testing.T) {
	mockService := &MockEventService{
		getUnsentEventsResponse: []eventmodel.Event{},
		getUnsentEventsError:    nil,
	}

	processor := NewDefaultOutboxProcessor(mockService)

	ctx := context.Background()
	processor.processEvents(ctx)

	if mockService.getUnsentEventsCalls != 1 {
		t.Errorf("expected 1 call to GetUnsentEvents, got %d", mockService.getUnsentEventsCalls)
	}
	if mockService.processEventCalls != 0 {
		t.Errorf("expected 0 calls to ProcessEvent, got %d", mockService.processEventCalls)
	}
}

func TestProcessor_processEvents_DatabaseError(t *testing.T) {
	expectedError := errors.New("database error")
	mockService := &MockEventService{
		getUnsentEventsResponse: []eventmodel.Event{},
		getUnsentEventsError:    expectedError,
	}

	processor := NewDefaultOutboxProcessor(mockService)

	ctx := context.Background()
	processor.processEvents(ctx)

	if mockService.getUnsentEventsCalls != 1 {
		t.Errorf("expected 1 call to GetUnsentEvents, got %d", mockService.getUnsentEventsCalls)
	}
	if mockService.processEventCalls != 0 {
		t.Errorf("expected 0 calls to ProcessEvent, got %d", mockService.processEventCalls)
	}
}

func TestProcessor_processEvents_SuccessfulProcessing(t *testing.T) {
	events := []eventmodel.Event{
		{
			ID:        uuid.New(),
			Name:      "test.event.1",
			Topic:     "test.topic",
			Payload:   []byte(`{"test": "data1"}`),
			Sent:      false,
			EventTime: time.Now(),
		},
		{
			ID:        uuid.New(),
			Name:      "test.event.2",
			Topic:     "test.topic",
			Payload:   []byte(`{"test": "data2"}`),
			Sent:      false,
			EventTime: time.Now(),
		},
	}

	mockService := &MockEventService{
		getUnsentEventsResponse: events,
		getUnsentEventsError:    nil,
		processEventError:       nil,
	}

	processor := NewDefaultOutboxProcessor(mockService)

	ctx := context.Background()
	processor.processEvents(ctx)

	if mockService.getUnsentEventsCalls != 1 {
		t.Errorf("expected 1 call to GetUnsentEvents, got %d", mockService.getUnsentEventsCalls)
	}
	if mockService.processEventCalls != 2 {
		t.Errorf("expected 2 calls to ProcessEvent, got %d", mockService.processEventCalls)
	}

	processedEvents := mockService.GetProcessedEvents()
	if len(processedEvents) != 2 {
		t.Errorf("expected 2 processed events, got %d", len(processedEvents))
	}
}

func TestProcessor_processEventWithRetry_SuccessFirstAttempt(t *testing.T) {
	event := eventmodel.Event{
		ID:        uuid.New(),
		Name:      "test.event",
		Topic:     "test.topic",
		Payload:   []byte(`{"test": "data"}`),
		Sent:      false,
		EventTime: time.Now(),
	}

	mockService := &MockEventService{
		processEventError: nil,
	}

	processor := NewDefaultOutboxProcessor(mockService)

	ctx := context.Background()
	err := processor.processEventWithRetry(ctx, event)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if mockService.GetProcessEventCallCount() != 1 {
		t.Errorf("expected 1 call to ProcessEvent, got %d", mockService.GetProcessEventCallCount())
	}
}

func TestProcessor_processEventWithRetry_SuccessAfterRetries(t *testing.T) {
	event := eventmodel.Event{
		ID:        uuid.New(),
		Name:      "test.event",
		Topic:     "test.topic",
		Payload:   []byte(`{"test": "data"}`),
		Sent:      false,
		EventTime: time.Now(),
	}

	// Fail twice, then succeed
	mockService := &MockEventService{}
	mockService.SetProcessEventCallSequence([]error{
		errors.New("temporary error"),
		errors.New("temporary error"),
		nil, // Success
	})

	config := Config{
		Interval:   30 * time.Second,
		BatchSize:  100,
		MaxRetries: 3,
		RetryDelay: 1 * time.Millisecond,
	}

	processor := NewOutboxProcessor(mockService, config)

	ctx := context.Background()
	err := processor.processEventWithRetry(ctx, event)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if mockService.GetProcessEventCallCount() != 3 {
		t.Errorf("expected 3 calls to ProcessEvent, got %d", mockService.GetProcessEventCallCount())
	}
}

func TestProcessor_processEventWithRetry_FailureAfterAllRetries(t *testing.T) {
	event := eventmodel.Event{
		ID:        uuid.New(),
		Name:      "test.event",
		Topic:     "test.topic",
		Payload:   []byte(`{"test": "data"}`),
		Sent:      false,
		EventTime: time.Now(),
	}

	expectedError := errors.New("persistent error")
	mockService := &MockEventService{
		processEventError: expectedError,
	}

	config := Config{
		Interval:   30 * time.Second,
		BatchSize:  100,
		MaxRetries: 3,
		RetryDelay: 1 * time.Millisecond,
	}

	processor := NewOutboxProcessor(mockService, config)

	ctx := context.Background()
	err := processor.processEventWithRetry(ctx, event)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if err != expectedError {
		t.Errorf("expected error %v, got %v", expectedError, err)
	}
	if mockService.GetProcessEventCallCount() != 3 {
		t.Errorf("expected 3 calls to ProcessEvent, got %d", mockService.GetProcessEventCallCount())
	}
}

func TestProcessor_processEventWithRetry_ContextCancellation(t *testing.T) {
	event := eventmodel.Event{
		ID:        uuid.New(),
		Name:      "test.event",
		Topic:     "test.topic",
		Payload:   []byte(`{"test": "data"}`),
		Sent:      false,
		EventTime: time.Now(),
	}

	mockService := &MockEventService{
		processEventError: errors.New("temporary error"),
	}

	config := Config{
		Interval:   30 * time.Second,
		BatchSize:  100,
		MaxRetries: 3,
		RetryDelay: 100 * time.Millisecond, // Longer delay to test cancellation
	}

	processor := NewOutboxProcessor(mockService, config)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := processor.processEventWithRetry(ctx, event)

	if err == nil {
		t.Error("expected error, got nil")
	}
	if err != context.DeadlineExceeded {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
	if mockService.GetProcessEventCallCount() != 1 {
		t.Errorf("expected 1 call to ProcessEvent, got %d", mockService.GetProcessEventCallCount())
	}
}

func TestProcessor_ProcessNow(t *testing.T) {
	events := []eventmodel.Event{
		{
			ID:        uuid.New(),
			Name:      "test.event",
			Topic:     "test.topic",
			Payload:   []byte(`{"test": "data"}`),
			Sent:      false,
			EventTime: time.Now(),
		},
	}

	mockService := &MockEventService{
		getUnsentEventsResponse: events,
		getUnsentEventsError:    nil,
		processEventError:       nil,
	}

	processor := NewDefaultOutboxProcessor(mockService)

	ctx := context.Background()
	err := processor.ProcessNow(ctx)

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if mockService.getUnsentEventsCalls != 1 {
		t.Errorf("expected 1 call to GetUnsentEvents, got %d", mockService.getUnsentEventsCalls)
	}
	if mockService.processEventCalls != 1 {
		t.Errorf("expected 1 call to ProcessEvent, got %d", mockService.processEventCalls)
	}
}

func TestProcessor_ProcessEvents_MixedSuccessAndFailure(t *testing.T) {
	event1 := eventmodel.Event{
		ID:        uuid.New(),
		Name:      "test.event.success",
		Topic:     "test.topic",
		Payload:   []byte(`{"test": "data1"}`),
		Sent:      false,
		EventTime: time.Now(),
	}

	event2 := eventmodel.Event{
		ID:        uuid.New(),
		Name:      "test.event.failure",
		Topic:     "test.topic",
		Payload:   []byte(`{"test": "data2"}`),
		Sent:      false,
		EventTime: time.Now(),
	}

	events := []eventmodel.Event{event1, event2}

	mockService := &MockEventService{
		getUnsentEventsResponse: events,
		getUnsentEventsError:    nil,
	}

	// Set first event to succeed, second to fail
	mockService.SetProcessEventErrorForEvent(event2.ID.String(), errors.New("processing error"))

	config := Config{
		Interval:   30 * time.Second,
		BatchSize:  100,
		MaxRetries: 2, // Reduced for faster test
		RetryDelay: 1 * time.Millisecond,
	}

	processor := NewOutboxProcessor(mockService, config)

	ctx := context.Background()
	processor.processEvents(ctx)

	if mockService.getUnsentEventsCalls != 1 {
		t.Errorf("expected 1 call to GetUnsentEvents, got %d", mockService.getUnsentEventsCalls)
	}

	// Should be 3 calls: 1 for event1 (success), 2 for event2 (fail with retry)
	expectedCalls := 1 + 2 // success + retries
	if mockService.processEventCalls != expectedCalls {
		t.Errorf("expected %d calls to ProcessEvent, got %d", expectedCalls, mockService.processEventCalls)
	}
}
