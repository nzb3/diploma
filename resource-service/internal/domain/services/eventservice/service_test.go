package eventservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/nzb3/diploma/resource-service/internal/domain/models/eventmodel"
)

// MockEventRepository implements the eventRepository interface for testing
type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) CreateEvent(ctx context.Context, event eventmodel.Event) (eventmodel.Event, error) {
	args := m.Called(ctx, event)
	return args.Get(0).(eventmodel.Event), args.Error(1)
}

func (m *MockEventRepository) GetNotSentEvents(ctx context.Context, limit int, offset int) ([]eventmodel.Event, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]eventmodel.Event), args.Error(1)
}

func (m *MockEventRepository) MarkEventAsSent(ctx context.Context, eventID uuid.UUID) error {
	args := m.Called(ctx, eventID)
	return args.Error(0)
}

// MockMessageProducer implements the messageProducer interface for testing
type MockMessageProducer struct {
	mock.Mock
}

func (m *MockMessageProducer) PublishEvent(ctx context.Context, event eventmodel.Event) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockMessageProducer) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// EventServiceTestSuite defines the test suite
type EventServiceTestSuite struct {
	suite.Suite
	service      *Service
	mockRepo     *MockEventRepository
	mockProducer *MockMessageProducer
	ctx          context.Context
	testEventID  uuid.UUID
	testEvent    eventmodel.Event
	testData     map[string]interface{}
}

// SetupTest runs before each test
func (suite *EventServiceTestSuite) SetupTest() {
	suite.mockRepo = &MockEventRepository{}
	suite.mockProducer = &MockMessageProducer{}
	suite.service = NewEventService(suite.mockRepo, suite.mockProducer)
	suite.ctx = context.Background()
	suite.testEventID = uuid.New()

	suite.testData = map[string]interface{}{
		"resource_id": "test-resource-id",
		"action":      "created",
		"timestamp":   time.Now().Unix(),
	}

	payload, _ := json.Marshal(suite.testData)
	suite.testEvent = eventmodel.Event{
		ID:        suite.testEventID,
		Name:      "resource.created",
		Topic:     "resources",
		Payload:   payload,
		Sent:      false,
		EventTime: time.Now(),
	}
}

// Test NewEventService
func (suite *EventServiceTestSuite) TestNewEventService() {
	service := NewEventService(suite.mockRepo, suite.mockProducer)

	assert.NotNil(suite.T(), service)
	assert.Equal(suite.T(), suite.mockRepo, service.eventRepo)
	assert.Equal(suite.T(), suite.mockProducer, service.producer)
}

// Test PublishEvent - Success case with immediate publish and mark as sent
func (suite *EventServiceTestSuite) TestPublishEvent_Success_ImmediatePublishAndMarkAsSent() {
	eventName := "resource.created"

	// Mock CreateEvent to return saved event
	savedEvent := suite.testEvent
	savedEvent.ID = uuid.New() // New ID assigned by repository
	suite.mockRepo.On("CreateEvent", suite.ctx, mock.MatchedBy(func(e eventmodel.Event) bool {
		return e.Name == eventName && e.Topic == "resources"
	})).Return(savedEvent, nil)

	// Mock successful publish
	suite.mockProducer.On("PublishEvent", suite.ctx, savedEvent).Return(nil)

	// Mock successful mark as sent
	suite.mockRepo.On("MarkEventAsSent", suite.ctx, savedEvent.ID).Return(nil)

	// Execute
	err := suite.service.PublishEvent(suite.ctx, eventName, suite.testData)

	// Assert
	assert.NoError(suite.T(), err)
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockProducer.AssertExpectations(suite.T())
}

// Test PublishEvent - CreateEvent fails
func (suite *EventServiceTestSuite) TestPublishEvent_CreateEventFails() {
	eventName := "resource.created"
	expectedError := errors.New("database error")

	suite.mockRepo.On("CreateEvent", suite.ctx, mock.MatchedBy(func(e eventmodel.Event) bool {
		return e.Name == eventName && e.Topic == "resources"
	})).Return(eventmodel.Event{}, expectedError)

	// Execute
	err := suite.service.PublishEvent(suite.ctx, eventName, suite.testData)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to save event to outbox")
	assert.Contains(suite.T(), err.Error(), expectedError.Error())
	suite.mockRepo.AssertExpectations(suite.T())
}

// Test PublishEvent - Publish fails but operation continues (outbox pattern)
func (suite *EventServiceTestSuite) TestPublishEvent_PublishFails_OutboxPattern() {
	eventName := "resource.created"
	savedEvent := suite.testEvent
	savedEvent.ID = uuid.New()
	publishError := errors.New("kafka connection failed")

	suite.mockRepo.On("CreateEvent", suite.ctx, mock.MatchedBy(func(e eventmodel.Event) bool {
		return e.Name == eventName && e.Topic == "resources"
	})).Return(savedEvent, nil)

	suite.mockProducer.On("PublishEvent", suite.ctx, savedEvent).Return(publishError)

	// Execute - should not fail even if publish fails (outbox pattern)
	err := suite.service.PublishEvent(suite.ctx, eventName, suite.testData)

	// Assert
	assert.NoError(suite.T(), err) // Should not fail due to outbox pattern
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockProducer.AssertExpectations(suite.T())
}

// Test PublishEvent - Publish succeeds but MarkEventAsSent fails
func (suite *EventServiceTestSuite) TestPublishEvent_MarkAsSentFails() {
	eventName := "resource.created"
	savedEvent := suite.testEvent
	savedEvent.ID = uuid.New()
	markSentError := errors.New("database connection lost")

	suite.mockRepo.On("CreateEvent", suite.ctx, mock.MatchedBy(func(e eventmodel.Event) bool {
		return e.Name == eventName && e.Topic == "resources"
	})).Return(savedEvent, nil)

	suite.mockProducer.On("PublishEvent", suite.ctx, savedEvent).Return(nil)
	suite.mockRepo.On("MarkEventAsSent", suite.ctx, savedEvent.ID).Return(markSentError)

	// Execute - should not fail even if marking as sent fails
	err := suite.service.PublishEvent(suite.ctx, eventName, suite.testData)

	// Assert
	assert.NoError(suite.T(), err) // Should not fail as event was published successfully
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockProducer.AssertExpectations(suite.T())
}

// Test PublishEvent - Invalid event data
func (suite *EventServiceTestSuite) TestPublishEvent_InvalidEventData() {
	eventName := "resource.created"
	// Use a channel as data, which cannot be JSON marshaled
	invalidData := make(chan int)

	// Execute
	err := suite.service.PublishEvent(suite.ctx, eventName, invalidData)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to create event")
}

// Test GetUnsentEvents - Success
func (suite *EventServiceTestSuite) TestGetUnsentEvents_Success() {
	limit, offset := 10, 0
	expectedEvents := []eventmodel.Event{suite.testEvent}

	suite.mockRepo.On("GetNotSentEvents", suite.ctx, limit, offset).Return(expectedEvents, nil)

	// Execute
	events, err := suite.service.GetUnsentEvents(suite.ctx, limit, offset)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedEvents, events)
	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetUnsentEvents - Repository error
func (suite *EventServiceTestSuite) TestGetUnsentEvents_RepositoryError() {
	limit, offset := 10, 0
	expectedError := errors.New("database error")

	suite.mockRepo.On("GetNotSentEvents", suite.ctx, limit, offset).Return([]eventmodel.Event{}, expectedError)

	// Execute
	events, err := suite.service.GetUnsentEvents(suite.ctx, limit, offset)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to retrieve unsent events")
	assert.Contains(suite.T(), err.Error(), expectedError.Error())
	assert.Nil(suite.T(), events)
	suite.mockRepo.AssertExpectations(suite.T())
}

// Test GetUnsentEvents - Empty result
func (suite *EventServiceTestSuite) TestGetUnsentEvents_EmptyResult() {
	limit, offset := 10, 0
	emptyEvents := []eventmodel.Event{}

	suite.mockRepo.On("GetNotSentEvents", suite.ctx, limit, offset).Return(emptyEvents, nil)

	// Execute
	events, err := suite.service.GetUnsentEvents(suite.ctx, limit, offset)

	// Assert
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), events)
	suite.mockRepo.AssertExpectations(suite.T())
}

// Test ProcessEvent - Success
func (suite *EventServiceTestSuite) TestProcessEvent_Success() {
	suite.mockProducer.On("PublishEvent", suite.ctx, suite.testEvent).Return(nil)
	suite.mockRepo.On("MarkEventAsSent", suite.ctx, suite.testEvent.ID).Return(nil)

	// Execute
	err := suite.service.ProcessEvent(suite.ctx, suite.testEvent)

	// Assert
	assert.NoError(suite.T(), err)
	suite.mockProducer.AssertExpectations(suite.T())
	suite.mockRepo.AssertExpectations(suite.T())
}

// Test ProcessEvent - Publish fails
func (suite *EventServiceTestSuite) TestProcessEvent_PublishFails() {
	publishError := errors.New("kafka broker unavailable")

	suite.mockProducer.On("PublishEvent", suite.ctx, suite.testEvent).Return(publishError)

	// Execute
	err := suite.service.ProcessEvent(suite.ctx, suite.testEvent)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to publish event")
	assert.Contains(suite.T(), err.Error(), publishError.Error())
	suite.mockProducer.AssertExpectations(suite.T())
}

// Test ProcessEvent - Publish succeeds but MarkEventAsSent fails
func (suite *EventServiceTestSuite) TestProcessEvent_MarkAsSentFails() {
	markSentError := errors.New("database connection lost")

	suite.mockProducer.On("PublishEvent", suite.ctx, suite.testEvent).Return(nil)
	suite.mockRepo.On("MarkEventAsSent", suite.ctx, suite.testEvent.ID).Return(markSentError)

	// Execute - should not fail even if marking as sent fails
	err := suite.service.ProcessEvent(suite.ctx, suite.testEvent)

	// Assert
	assert.NoError(suite.T(), err) // Should not fail as event was published successfully
	suite.mockProducer.AssertExpectations(suite.T())
	suite.mockRepo.AssertExpectations(suite.T())
}

// Test Health - Success
func (suite *EventServiceTestSuite) TestHealth_Success() {
	suite.mockProducer.On("Health", suite.ctx).Return(nil)

	// Execute
	err := suite.service.Health(suite.ctx)

	// Assert
	assert.NoError(suite.T(), err)
	suite.mockProducer.AssertExpectations(suite.T())
}

// Test Health - Producer health check fails
func (suite *EventServiceTestSuite) TestHealth_ProducerHealthFails() {
	healthError := errors.New("producer is unhealthy")

	suite.mockProducer.On("Health", suite.ctx).Return(healthError)

	// Execute
	err := suite.service.Health(suite.ctx)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "message producer health check failed")
	assert.Contains(suite.T(), err.Error(), healthError.Error())
	suite.mockProducer.AssertExpectations(suite.T())
}

// Test with different event names and topics
func (suite *EventServiceTestSuite) TestPublishEvent_DifferentEventTypes() {
	testCases := []struct {
		name      string
		eventName string
		data      interface{}
	}{
		{
			name:      "resource.created",
			eventName: "resource.created",
			data:      map[string]interface{}{"id": "123", "name": "test"},
		},
		{
			name:      "resource.updated",
			eventName: "resource.updated",
			data:      map[string]interface{}{"id": "456", "changes": []string{"name"}},
		},
		{
			name:      "resource.deleted",
			eventName: "resource.deleted",
			data:      map[string]interface{}{"id": "789"},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Setup fresh mocks for each test case
			suite.SetupTest()

			savedEvent := suite.testEvent
			savedEvent.ID = uuid.New()
			savedEvent.Name = tc.eventName

			suite.mockRepo.On("CreateEvent", suite.ctx, mock.MatchedBy(func(e eventmodel.Event) bool {
				return e.Name == tc.eventName && e.Topic == "resources"
			})).Return(savedEvent, nil)

			suite.mockProducer.On("PublishEvent", suite.ctx, savedEvent).Return(nil)
			suite.mockRepo.On("MarkEventAsSent", suite.ctx, savedEvent.ID).Return(nil)

			// Execute
			err := suite.service.PublishEvent(suite.ctx, tc.eventName, tc.data)

			// Assert
			assert.NoError(suite.T(), err)
			suite.mockRepo.AssertExpectations(suite.T())
			suite.mockProducer.AssertExpectations(suite.T())
		})
	}
}

// Test context cancellation
func (suite *EventServiceTestSuite) TestPublishEvent_ContextCancelled() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	eventName := "resource.created"

	// Even with cancelled context, the operation should depend on repository behavior
	suite.mockRepo.On("CreateEvent", ctx, mock.MatchedBy(func(e eventmodel.Event) bool {
		return e.Name == eventName && e.Topic == "resources"
	})).Return(eventmodel.Event{}, context.Canceled)

	// Execute
	err := suite.service.PublishEvent(ctx, eventName, suite.testData)

	// Assert
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to save event to outbox")
	suite.mockRepo.AssertExpectations(suite.T())
}

// Test concurrent access safety (basic concurrency test)
func (suite *EventServiceTestSuite) TestPublishEvent_Concurrency() {
	// This test ensures the service can handle concurrent calls without panic
	// In a real scenario, you'd use more sophisticated concurrency testing

	numGoroutines := 10
	eventName := "resource.created"

	for i := 0; i < numGoroutines; i++ {
		savedEvent := suite.testEvent
		savedEvent.ID = uuid.New()

		suite.mockRepo.On("CreateEvent", suite.ctx, mock.MatchedBy(func(e eventmodel.Event) bool {
			return e.Name == eventName && e.Topic == "resources"
		})).Return(savedEvent, nil).Maybe()

		suite.mockProducer.On("PublishEvent", suite.ctx, savedEvent).Return(nil).Maybe()
		suite.mockRepo.On("MarkEventAsSent", suite.ctx, savedEvent.ID).Return(nil).Maybe()
	}

	// Execute concurrently
	errChan := make(chan error, numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			err := suite.service.PublishEvent(suite.ctx, eventName, suite.testData)
			errChan <- err
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-errChan
		assert.NoError(suite.T(), err)
	}
}

// Run the test suite
func TestEventServiceTestSuite(t *testing.T) {
	suite.Run(t, new(EventServiceTestSuite))
}

// Additional individual tests for edge cases
func TestNewEventService_NilInputs(t *testing.T) {
	service := NewEventService(nil, nil)
	assert.NotNil(t, service)
	assert.Nil(t, service.eventRepo)
	assert.Nil(t, service.producer)
}

func TestService_PublishEvent_EmptyEventName(t *testing.T) {
	mockRepo := &MockEventRepository{}
	mockProducer := &MockMessageProducer{}
	service := NewEventService(mockRepo, mockProducer)

	data := map[string]interface{}{"test": "data"}

	// Create event with empty name should still work
	savedEvent := eventmodel.Event{
		ID:    uuid.New(),
		Name:  "", // Empty name
		Topic: "resources",
	}

	mockRepo.On("CreateEvent", mock.Anything, mock.MatchedBy(func(e eventmodel.Event) bool {
		return e.Name == "" && e.Topic == "resources"
	})).Return(savedEvent, nil)

	mockProducer.On("PublishEvent", mock.Anything, savedEvent).Return(nil)
	mockRepo.On("MarkEventAsSent", mock.Anything, savedEvent.ID).Return(nil)

	err := service.PublishEvent(context.Background(), "", data)
	assert.NoError(t, err)
}

func TestService_GetUnsentEvents_WithPagination(t *testing.T) {
	mockRepo := &MockEventRepository{}
	mockProducer := &MockMessageProducer{}
	service := NewEventService(mockRepo, mockProducer)

	// Test different pagination scenarios
	testCases := []struct {
		limit  int
		offset int
	}{
		{10, 0},  // First page
		{10, 10}, // Second page
		{5, 15},  // Smaller page size
		{100, 0}, // Large page size
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("limit_%d_offset_%d", tc.limit, tc.offset), func(t *testing.T) {
			expectedEvents := []eventmodel.Event{{ID: uuid.New()}}
			mockRepo.On("GetNotSentEvents", mock.Anything, tc.limit, tc.offset).Return(expectedEvents, nil).Once()

			events, err := service.GetUnsentEvents(context.Background(), tc.limit, tc.offset)

			assert.NoError(t, err)
			assert.Equal(t, expectedEvents, events)
		})
	}
}
