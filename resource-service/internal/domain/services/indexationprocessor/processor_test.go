package indexationprocessor

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/nzb3/diploma/resource-service/internal/domain/models/resourcemodel"
	"github.com/nzb3/diploma/resource-service/internal/repository/messaging"
)

// MockResourceService is a mock implementation of resourceService interface
type MockResourceService struct {
	mock.Mock
}

func (m *MockResourceService) UpdateResourceStatus(ctx context.Context, resource resourcemodel.Resource, status resourcemodel.ResourceStatus) (resourcemodel.Resource, error) {
	args := m.Called(ctx, resource, status)
	return args.Get(0).(resourcemodel.Resource), args.Error(1)
}

func (m *MockResourceService) GetResourceStatusChannel(resourceID uuid.UUID) (chan resourcemodel.ResourceStatusUpdate, bool) {
	args := m.Called(resourceID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(chan resourcemodel.ResourceStatusUpdate), args.Bool(1)
}

func (m *MockResourceService) RemoveResourceStatusChannel(resourceID uuid.UUID) {
	m.Called(resourceID)
}

func (m *MockResourceService) GetResourceByID(ctx context.Context, resourceID uuid.UUID) (resourcemodel.Resource, error) {
	args := m.Called(ctx, resourceID)
	return args.Get(0).(resourcemodel.Resource), args.Error(1)
}

// MockMessageConsumer is a mock implementation of messaging.MessageConsumer interface
type MockMessageConsumer struct {
	mock.Mock
}

func (m *MockMessageConsumer) Subscribe(ctx context.Context, topics []string, handler messaging.MessageHandler) error {
	args := m.Called(ctx, topics, handler)
	return args.Error(0)
}

func (m *MockMessageConsumer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockMessageConsumer) Health(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// IndexationProcessorTestSuite is the test suite for IndexationProcessor
type IndexationProcessorTestSuite struct {
	suite.Suite
	mockResourceService *MockResourceService
	mockConsumer        *MockMessageConsumer
	processor           *Processor
	ctx                 context.Context
}

func (suite *IndexationProcessorTestSuite) SetupTest() {
	suite.mockResourceService = new(MockResourceService)
	suite.mockConsumer = new(MockMessageConsumer)
	suite.processor = NewIndexationProcessor(suite.mockResourceService, suite.mockConsumer)
	suite.ctx = context.Background()
}

func (suite *IndexationProcessorTestSuite) TearDownTest() {
	suite.mockResourceService.AssertExpectations(suite.T())
	suite.mockConsumer.AssertExpectations(suite.T())
}

// TestNewIndexationProcessor tests the constructor
func (suite *IndexationProcessorTestSuite) TestNewIndexationProcessor() {
	processor := NewIndexationProcessor(suite.mockResourceService, suite.mockConsumer)

	assert.NotNil(suite.T(), processor)
	assert.Equal(suite.T(), suite.mockResourceService, processor.resourceService)
	assert.Equal(suite.T(), suite.mockConsumer, processor.consumer)
	assert.NotNil(suite.T(), processor.stopCh)
	assert.NotNil(suite.T(), processor.doneCh)
}

// TestStart_Success tests successful start of the processor
func (suite *IndexationProcessorTestSuite) TestStart_Success() {
	topics := []string{"indexation_complete"}
	
	suite.mockConsumer.On("Subscribe", mock.Anything, topics, suite.processor).Return(nil).Once()
	
	// Create a context that will be cancelled after a short time to stop the processor
	ctx, cancel := context.WithTimeout(suite.ctx, 100*time.Millisecond)
	defer cancel()
	
	err := suite.processor.Start(ctx)
	
	assert.NoError(suite.T(), err)
}

// TestStart_SubscribeError tests start failure due to subscription error
func (suite *IndexationProcessorTestSuite) TestStart_SubscribeError() {
	topics := []string{"indexation_complete"}
	expectedError := errors.New("subscription failed")
	
	suite.mockConsumer.On("Subscribe", mock.Anything, topics, suite.processor).Return(expectedError).Once()
	
	err := suite.processor.Start(suite.ctx)
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to subscribe to topics")
	assert.Contains(suite.T(), err.Error(), expectedError.Error())
}

// TestStop tests graceful stopping of the processor
func (suite *IndexationProcessorTestSuite) TestStop() {
	suite.mockConsumer.On("Close").Return(nil).Once()
	
	// Start the processor in a goroutine
	go func() {
		time.Sleep(50 * time.Millisecond)
		suite.processor.Stop()
	}()
	
	// Start should return when Stop is called
	ctx, cancel := context.WithTimeout(suite.ctx, 200*time.Millisecond)
	defer cancel()
	
	suite.mockConsumer.On("Subscribe", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	
	err := suite.processor.Start(ctx)
	assert.NoError(suite.T(), err)
}

// TestHandleMessage_Success tests successful message handling
func (suite *IndexationProcessorTestSuite) TestHandleMessage_Success() {
	resourceID := uuid.New()
	event := IndexationCompleteEvent{
		ResourceID: resourceID,
		Success:    true,
		Message:    "Indexation completed successfully",
	}
	
	eventJSON, _ := json.Marshal(event)
	
	resource := resourcemodel.Resource{
		ID:     resourceID,
		Status: resourcemodel.ResourceStatusProcessing,
	}
	
	updatedResource := resource
	updatedResource.Status = resourcemodel.ResourceStatusCompleted
	
	statusCh := make(chan resourcemodel.ResourceStatusUpdate, 1)
	
	// Setup expectations
	suite.mockResourceService.On("GetResourceByID", mock.Anything, resourceID).Return(resource, nil).Once()
	suite.mockResourceService.On("UpdateResourceStatus", mock.Anything, resource, resourcemodel.ResourceStatusCompleted).Return(updatedResource, nil).Once()
	suite.mockResourceService.On("GetResourceStatusChannel", resourceID).Return(statusCh, true).Once()
	suite.mockResourceService.On("RemoveResourceStatusChannel", resourceID).Once()
	
	err := suite.processor.HandleMessage(suite.ctx, "indexation_complete", resourceID.String(), eventJSON, nil)
	
	assert.NoError(suite.T(), err)
	
	// Verify that status update was sent to channel
	select {
	case statusUpdate := <-statusCh:
		assert.Equal(suite.T(), resourceID, statusUpdate.ResourceID)
		assert.Equal(suite.T(), resourcemodel.ResourceStatusCompleted, statusUpdate.Status)
	case <-time.After(100 * time.Millisecond):
		suite.T().Fatal("Expected status update not received")
	}
	
	// Verify channel is closed
	_, ok := <-statusCh
	assert.False(suite.T(), ok, "Channel should be closed")
}

// TestHandleMessage_FailedIndexation tests handling failed indexation event
func (suite *IndexationProcessorTestSuite) TestHandleMessage_FailedIndexation() {
	resourceID := uuid.New()
	event := IndexationCompleteEvent{
		ResourceID: resourceID,
		Success:    false,
		Message:    "Indexation failed",
	}
	
	eventJSON, _ := json.Marshal(event)
	
	resource := resourcemodel.Resource{
		ID:     resourceID,
		Status: resourcemodel.ResourceStatusProcessing,
	}
	
	updatedResource := resource
	updatedResource.Status = resourcemodel.ResourceStatusFailed
	
	statusCh := make(chan resourcemodel.ResourceStatusUpdate, 1)
	
	// Setup expectations
	suite.mockResourceService.On("GetResourceByID", mock.Anything, resourceID).Return(resource, nil).Once()
	suite.mockResourceService.On("UpdateResourceStatus", mock.Anything, resource, resourcemodel.ResourceStatusFailed).Return(updatedResource, nil).Once()
	suite.mockResourceService.On("GetResourceStatusChannel", resourceID).Return(statusCh, true).Once()
	suite.mockResourceService.On("RemoveResourceStatusChannel", resourceID).Once()
	
	err := suite.processor.HandleMessage(suite.ctx, "indexation_complete", resourceID.String(), eventJSON, nil)
	
	assert.NoError(suite.T(), err)
	
	// Verify that status update was sent to channel
	select {
	case statusUpdate := <-statusCh:
		assert.Equal(suite.T(), resourceID, statusUpdate.ResourceID)
		assert.Equal(suite.T(), resourcemodel.ResourceStatusFailed, statusUpdate.Status)
	case <-time.After(100 * time.Millisecond):
		suite.T().Fatal("Expected status update not received")
	}
}

// TestHandleMessage_InvalidJSON tests handling invalid JSON payload
func (suite *IndexationProcessorTestSuite) TestHandleMessage_InvalidJSON() {
	resourceID := uuid.New()
	invalidJSON := []byte(`{"invalid": "json"`)
	
	err := suite.processor.HandleMessage(suite.ctx, "indexation_complete", resourceID.String(), invalidJSON, nil)
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to unmarshal event")
}

// TestHandleMessage_GetResourceError tests handling error when getting resource
func (suite *IndexationProcessorTestSuite) TestHandleMessage_GetResourceError() {
	resourceID := uuid.New()
	event := IndexationCompleteEvent{
		ResourceID: resourceID,
		Success:    true,
		Message:    "Indexation completed successfully",
	}
	
	eventJSON, _ := json.Marshal(event)
	expectedError := errors.New("resource not found")
	
	suite.mockResourceService.On("GetResourceByID", mock.Anything, resourceID).Return(resourcemodel.Resource{}, expectedError).Once()
	
	err := suite.processor.HandleMessage(suite.ctx, "indexation_complete", resourceID.String(), eventJSON, nil)
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to get resource")
	assert.Contains(suite.T(), err.Error(), expectedError.Error())
}

// TestHandleMessage_UpdateStatusError tests handling error when updating resource status
func (suite *IndexationProcessorTestSuite) TestHandleMessage_UpdateStatusError() {
	resourceID := uuid.New()
	event := IndexationCompleteEvent{
		ResourceID: resourceID,
		Success:    true,
		Message:    "Indexation completed successfully",
	}
	
	eventJSON, _ := json.Marshal(event)
	
	resource := resourcemodel.Resource{
		ID:     resourceID,
		Status: resourcemodel.ResourceStatusProcessing,
	}
	
	expectedError := errors.New("update failed")
	
	suite.mockResourceService.On("GetResourceByID", mock.Anything, resourceID).Return(resource, nil).Once()
	suite.mockResourceService.On("UpdateResourceStatus", mock.Anything, resource, resourcemodel.ResourceStatusCompleted).Return(resourcemodel.Resource{}, expectedError).Once()
	
	err := suite.processor.HandleMessage(suite.ctx, "indexation_complete", resourceID.String(), eventJSON, nil)
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to update resource status")
	assert.Contains(suite.T(), err.Error(), expectedError.Error())
}

// TestHandleMessage_NoStatusChannel tests handling when no status channel exists
func (suite *IndexationProcessorTestSuite) TestHandleMessage_NoStatusChannel() {
	resourceID := uuid.New()
	event := IndexationCompleteEvent{
		ResourceID: resourceID,
		Success:    true,
		Message:    "Indexation completed successfully",
	}
	
	eventJSON, _ := json.Marshal(event)
	
	resource := resourcemodel.Resource{
		ID:     resourceID,
		Status: resourcemodel.ResourceStatusProcessing,
	}
	
	updatedResource := resource
	updatedResource.Status = resourcemodel.ResourceStatusCompleted
	
	// Setup expectations - no status channel exists
	suite.mockResourceService.On("GetResourceByID", mock.Anything, resourceID).Return(resource, nil).Once()
	suite.mockResourceService.On("UpdateResourceStatus", mock.Anything, resource, resourcemodel.ResourceStatusCompleted).Return(updatedResource, nil).Once()
	suite.mockResourceService.On("GetResourceStatusChannel", resourceID).Return(nil, false).Once()
	
	err := suite.processor.HandleMessage(suite.ctx, "indexation_complete", resourceID.String(), eventJSON, nil)
	
	assert.NoError(suite.T(), err)
}

// TestHandleMessage_FullStatusChannel tests handling when status channel is full
func (suite *IndexationProcessorTestSuite) TestHandleMessage_FullStatusChannel() {
	resourceID := uuid.New()
	event := IndexationCompleteEvent{
		ResourceID: resourceID,
		Success:    true,
		Message:    "Indexation completed successfully",
	}
	
	eventJSON, _ := json.Marshal(event)
	
	resource := resourcemodel.Resource{
		ID:     resourceID,
		Status: resourcemodel.ResourceStatusProcessing,
	}
	
	updatedResource := resource
	updatedResource.Status = resourcemodel.ResourceStatusCompleted
	
	// Create a channel with no buffer (will be full immediately)
	statusCh := make(chan resourcemodel.ResourceStatusUpdate)
	
	// Setup expectations
	suite.mockResourceService.On("GetResourceByID", mock.Anything, resourceID).Return(resource, nil).Once()
	suite.mockResourceService.On("UpdateResourceStatus", mock.Anything, resource, resourcemodel.ResourceStatusCompleted).Return(updatedResource, nil).Once()
	suite.mockResourceService.On("GetResourceStatusChannel", resourceID).Return(statusCh, true).Once()
	suite.mockResourceService.On("RemoveResourceStatusChannel", resourceID).Once()
	
	err := suite.processor.HandleMessage(suite.ctx, "indexation_complete", resourceID.String(), eventJSON, nil)
	
	assert.NoError(suite.T(), err)
	
	// Verify channel is closed
	_, ok := <-statusCh
	assert.False(suite.T(), ok, "Channel should be closed")
}

// TestHandleMessage_ContextCancellation tests handling context cancellation during status update
func (suite *IndexationProcessorTestSuite) TestHandleMessage_ContextCancellation() {
	resourceID := uuid.New()
	event := IndexationCompleteEvent{
		ResourceID: resourceID,
		Success:    true,
		Message:    "Indexation completed successfully",
	}
	
	eventJSON, _ := json.Marshal(event)
	
	resource := resourcemodel.Resource{
		ID:     resourceID,
		Status: resourcemodel.ResourceStatusProcessing,
	}
	
	updatedResource := resource
	updatedResource.Status = resourcemodel.ResourceStatusCompleted
	
	statusCh := make(chan resourcemodel.ResourceStatusUpdate)
	
	// Create a context that will be cancelled immediately
	ctx, cancel := context.WithCancel(suite.ctx)
	cancel()
	
	// Setup expectations
	suite.mockResourceService.On("GetResourceByID", mock.Anything, resourceID).Return(resource, nil).Once()
	suite.mockResourceService.On("UpdateResourceStatus", mock.Anything, resource, resourcemodel.ResourceStatusCompleted).Return(updatedResource, nil).Once()
	suite.mockResourceService.On("GetResourceStatusChannel", resourceID).Return(statusCh, true).Once()
	
	err := suite.processor.HandleMessage(ctx, "indexation_complete", resourceID.String(), eventJSON, nil)
	
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), context.Canceled, err)
}

// TestHandleMessage_IgnoreOtherTopics tests that messages from other topics are ignored
func (suite *IndexationProcessorTestSuite) TestHandleMessage_IgnoreOtherTopics() {
	resourceID := uuid.New()
	
	err := suite.processor.HandleMessage(suite.ctx, "other_topic", resourceID.String(), []byte("some data"), nil)
	
	assert.NoError(suite.T(), err)
	// No expectations should be called since the topic is ignored
}

// TestHealth_Success tests successful health check
func (suite *IndexationProcessorTestSuite) TestHealth_Success() {
	suite.mockConsumer.On("Health", mock.Anything).Return(nil).Once()
	
	err := suite.processor.Health(suite.ctx)
	
	assert.NoError(suite.T(), err)
}

// TestHealth_ConsumerError tests health check failure
func (suite *IndexationProcessorTestSuite) TestHealth_ConsumerError() {
	expectedError := errors.New("consumer unhealthy")
	suite.mockConsumer.On("Health", mock.Anything).Return(expectedError).Once()
	
	err := suite.processor.Health(suite.ctx)
	
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
}

// TestHealth_NoConsumer tests health check when consumer is nil
func (suite *IndexationProcessorTestSuite) TestHealth_NoConsumer() {
	processor := &Processor{
		resourceService: suite.mockResourceService,
		consumer:        nil,
		stopCh:          make(chan struct{}),
		doneCh:          make(chan struct{}),
	}
	
	err := processor.Health(suite.ctx)
	
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "consumer not initialized")
}

// TestIndexationCompleteEvent_JSONMarshalUnmarshal tests JSON marshaling/unmarshaling of IndexationCompleteEvent
func (suite *IndexationProcessorTestSuite) TestIndexationCompleteEvent_JSONMarshalUnmarshal() {
	resourceID := uuid.New()
	original := IndexationCompleteEvent{
		ResourceID: resourceID,
		Success:    true,
		Message:    "Test message",
	}
	
	// Marshal to JSON
	data, err := json.Marshal(original)
	assert.NoError(suite.T(), err)
	
	// Unmarshal from JSON
	var unmarshaled IndexationCompleteEvent
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(suite.T(), err)
	
	// Verify fields
	assert.Equal(suite.T(), original.ResourceID, unmarshaled.ResourceID)
	assert.Equal(suite.T(), original.Success, unmarshaled.Success)
	assert.Equal(suite.T(), original.Message, unmarshaled.Message)
}

// Run the test suite
func TestIndexationProcessorSuite(t *testing.T) {
	suite.Run(t, new(IndexationProcessorTestSuite))
}

// Individual test functions for additional coverage

func TestNewIndexationProcessor(t *testing.T) {
	mockResourceService := new(MockResourceService)
	mockConsumer := new(MockMessageConsumer)
	
	processor := NewIndexationProcessor(mockResourceService, mockConsumer)
	
	assert.NotNil(t, processor)
	assert.Equal(t, mockResourceService, processor.resourceService)
	assert.Equal(t, mockConsumer, processor.consumer)
	assert.NotNil(t, processor.stopCh)
	assert.NotNil(t, processor.doneCh)
}

func TestIndexationCompleteEvent_EmptyMessage(t *testing.T) {
	resourceID := uuid.New()
	event := IndexationCompleteEvent{
		ResourceID: resourceID,
		Success:    false,
		// Message is omitted - should be empty string
	}
	
	data, err := json.Marshal(event)
	assert.NoError(t, err)
	
	var unmarshaled IndexationCompleteEvent
	err = json.Unmarshal(data, &unmarshaled)
	assert.NoError(t, err)
	
	assert.Equal(t, resourceID, unmarshaled.ResourceID)
	assert.False(t, unmarshaled.Success)
	assert.Empty(t, unmarshaled.Message)
}