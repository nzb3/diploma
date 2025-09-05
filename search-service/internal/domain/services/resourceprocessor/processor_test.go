package resourceprocessor
package resourceprocessor

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

	"github.com/nzb3/diploma/search-service/internal/domain/models"
	"github.com/nzb3/diploma/search-service/internal/repository/messaging"
)

// MockVectorStorage is a mock implementation of vectorStorage interface
type MockVectorStorage struct {
	mock.Mock
}

func (m *MockVectorStorage) PutResource(ctx context.Context, resource models.Resource) ([]string, error) {
	args := m.Called(ctx, resource)
	return args.Get(0).([]string), args.Error(1)
}

// MockEventService is a mock implementation of eventService interface
type MockEventService struct {
	mock.Mock
}

func (m *MockEventService) PublishEvent(ctx context.Context, topic string, eventName string, data interface{}) error {
	args := m.Called(ctx, topic, eventName, data)
	return args.Error(0)
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

// ResourceProcessorTestSuite is the test suite for ResourceProcessor
type ResourceProcessorTestSuite struct {
	suite.Suite
	mockVectorStorage *MockVectorStorage
	mockEventService  *MockEventService
	mockConsumer      *MockMessageConsumer
	processor         *Processor
	ctx               context.Context
}

func (suite *ResourceProcessorTestSuite) SetupTest() {
	suite.mockVectorStorage = new(MockVectorStorage)
	suite.mockEventService = new(MockEventService)
	suite.mockConsumer = new(MockMessageConsumer)
	suite.processor = NewResourceProcessor(suite.mockVectorStorage, suite.mockEventService, suite.mockConsumer)
	suite.ctx = context.Background()
}

func (suite *ResourceProcessorTestSuite) TearDownTest() {
	suite.mockVectorStorage.AssertExpectations(suite.T())
	suite.mockEventService.AssertExpectations(suite.T())
	suite.mockConsumer.AssertExpectations(suite.T())
}

// TestNewResourceProcessor tests the constructor
func (suite *ResourceProcessorTestSuite) TestNewResourceProcessor() {
	processor := NewResourceProcessor(suite.mockVectorStorage, suite.mockEventService, suite.mockConsumer)

	assert.NotNil(suite.T(), processor)
	assert.Equal(suite.T(), suite.mockVectorStorage, processor.vectorStorage)
	assert.Equal(suite.T(), suite.mockEventService, processor.eventService)
	assert.Equal(suite.T(), suite.mockConsumer, processor.consumer)
	assert.NotNil(suite.T(), processor.stopCh)
	assert.NotNil(suite.T(), processor.doneCh)
}

// TestStart_Success tests successful start of the processor
func (suite *ResourceProcessorTestSuite) TestStart_Success() {
	topics := []string{\"resource\"}

	suite.mockConsumer.On(\"Subscribe\", mock.Anything, topics, suite.processor).Return(nil).Once()

	// Create a context that will be cancelled after a short time to stop the processor
	ctx, cancel := context.WithTimeout(suite.ctx, 100*time.Millisecond)
	defer cancel()

	err := suite.processor.Start(ctx)

	assert.NoError(suite.T(), err)
}

// TestStart_SubscribeError tests start failure due to subscription error
func (suite *ResourceProcessorTestSuite) TestStart_SubscribeError() {
	topics := []string{\"resource\"}
	expectedError := errors.New(\"subscription failed\")

	suite.mockConsumer.On(\"Subscribe\", mock.Anything, topics, suite.processor).Return(expectedError).Once()

	err := suite.processor.Start(suite.ctx)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), \"failed to subscribe to topics\")
	assert.Contains(suite.T(), err.Error(), expectedError.Error())
}

// TestHandleMessage_Success tests successful message handling
func (suite *ResourceProcessorTestSuite) TestHandleMessage_Success() {
	resourceID := uuid.New()
	resource := models.Resource{
		ID:               resourceID,
		Name:             \"test-resource\",
		Type:             \"text\",
		ExtractedContent: \"test content\",
	}

	resourceJSON, _ := json.Marshal(resource)
	headers := map[string]string{
		\"event-name\": \"resource.created\",
	}

	chunkIDs := []string{\"chunk1\", \"chunk2\"}

	expectedEvent := IndexationCompleteEvent{
		ResourceID: resourceID,
		Success:    true,
		Message:    \"Resource indexed successfully\",
		ChunkIDs:   chunkIDs,
	}

	// Setup expectations
	suite.mockVectorStorage.On(\"PutResource\", mock.Anything, resource).Return(chunkIDs, nil).Once()
	suite.mockEventService.On(\"PublishEvent\", mock.Anything, \"indexation_complete\", \"indexation_complete\", expectedEvent).Return(nil).Once()

	err := suite.processor.HandleMessage(suite.ctx, \"resource\", resourceID.String(), resourceJSON, headers)

	assert.NoError(suite.T(), err)
}

// TestHandleMessage_VectorStorageError tests handling vector storage error
func (suite *ResourceProcessorTestSuite) TestHandleMessage_VectorStorageError() {
	resourceID := uuid.New()
	resource := models.Resource{
		ID:               resourceID,
		Name:             \"test-resource\",
		Type:             \"text\",
		ExtractedContent: \"test content\",
	}

	resourceJSON, _ := json.Marshal(resource)
	headers := map[string]string{
		\"event-name\": \"resource.created\",
	}

	expectedError := errors.New(\"vector storage error\")

	expectedEvent := IndexationCompleteEvent{
		ResourceID: resourceID,
		Success:    false,
		Message:    expectedError.Error(),
		ChunkIDs:   nil,
	}

	// Setup expectations
	suite.mockVectorStorage.On(\"PutResource\", mock.Anything, resource).Return([]string{}, expectedError).Once()
	suite.mockEventService.On(\"PublishEvent\", mock.Anything, \"indexation_complete\", \"indexation_complete\", expectedEvent).Return(nil).Once()

	err := suite.processor.HandleMessage(suite.ctx, \"resource\", resourceID.String(), resourceJSON, headers)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), \"failed to process resource\")
}

// TestHandleMessage_InvalidJSON tests handling invalid JSON payload
func (suite *ResourceProcessorTestSuite) TestHandleMessage_InvalidJSON() {
	resourceID := uuid.New()
	invalidJSON := []byte(`{\"invalid\": \"json\"`)
	headers := map[string]string{
		\"event-name\": \"resource.created\",
	}

	err := suite.processor.HandleMessage(suite.ctx, \"resource\", resourceID.String(), invalidJSON, headers)

	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), \"failed to unmarshal resource\")
}

// TestHandleMessage_IgnoreOtherTopics tests that messages from other topics are ignored
func (suite *ResourceProcessorTestSuite) TestHandleMessage_IgnoreOtherTopics() {
	resourceID := uuid.New()

	err := suite.processor.HandleMessage(suite.ctx, \"other_topic\", resourceID.String(), []byte(\"some data\"), nil)

	assert.NoError(suite.T(), err)
	// No expectations should be called since the topic is ignored
}

// TestHandleMessage_IgnoreOtherEvents tests that non-resource.created events are ignored
func (suite *ResourceProcessorTestSuite) TestHandleMessage_IgnoreOtherEvents() {
	resourceID := uuid.New()
	headers := map[string]string{
		\"event-name\": \"resource.updated\",
	}

	err := suite.processor.HandleMessage(suite.ctx, \"resource\", resourceID.String(), []byte(\"some data\"), headers)

	assert.NoError(suite.T(), err)
	// No expectations should be called since the event is ignored
}

// TestHandleMessage_MissingEventName tests handling missing event-name header
func (suite *ResourceProcessorTestSuite) TestHandleMessage_MissingEventName() {
	resourceID := uuid.New()
	headers := map[string]string{} // No event-name header

	err := suite.processor.HandleMessage(suite.ctx, \"resource\", resourceID.String(), []byte(\"some data\"), headers)

	assert.NoError(suite.T(), err)
	// No expectations should be called since the event-name is missing
}

// TestHealth_Success tests successful health check
func (suite *ResourceProcessorTestSuite) TestHealth_Success() {
	suite.mockConsumer.On(\"Health\", mock.Anything).Return(nil).Once()

	err := suite.processor.Health(suite.ctx)

	assert.NoError(suite.T(), err)
}

// TestHealth_ConsumerError tests health check failure
func (suite *ResourceProcessorTestSuite) TestHealth_ConsumerError() {
	expectedError := errors.New(\"consumer unhealthy\")
	suite.mockConsumer.On(\"Health\", mock.Anything).Return(expectedError).Once()

	err := suite.processor.Health(suite.ctx)

	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), expectedError, err)
}

func TestResourceProcessorTestSuite(t *testing.T) {
	suite.Run(t, new(ResourceProcessorTestSuite))
}