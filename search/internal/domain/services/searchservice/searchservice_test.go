package searchservice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/nzb3/diploma/search/internal/domain/models"
)

// Mock for vectorStorage
type mockVectorStorage struct {
	mock.Mock
}

func (m *mockVectorStorage) GetAnswer(ctx context.Context, question string, refsCh chan<- []models.Reference) (models.SearchResult, error) {
	args := m.Called(ctx, question, refsCh)
	return args.Get(0).(models.SearchResult), args.Error(1)
}

func (m *mockVectorStorage) GetAnswerStream(ctx context.Context, question string) (<-chan models.SearchResult, <-chan []models.Reference, <-chan []byte, <-chan error) {
	args := m.Called(ctx, question, refsCh, chunkCh)
	return nil, nil, nil, args.Error(1)
}

func (m *mockVectorStorage) SemanticSearch(ctx context.Context, query string) ([]models.Reference, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]models.Reference), args.Error(1)
}

// Mock for repository
type mockRepository struct {
	mock.Mock
}

func (m *mockRepository) GetResourceIDByReference(ctx context.Context, reference models.Reference) (uuid.UUID, error) {
	args := m.Called(ctx, reference)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func TestService_GetAnswerStream(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		question       string
		setupMocks     func(*mockVectorStorage, *mockRepository)
		mockReferences []models.Reference
		expectedResult models.SearchResult
		expectedErr    string
		contextCancel  bool
	}{
		{
			name:     "Success case",
			question: "What is the meaning of life?",
			setupMocks: func(vs *mockVectorStorage, repo *mockRepository) {
				// Setup references that will be "returned" by vector storage
				refs := []models.Reference{
					{Content: "Life has meaning", Score: 0.95},
					{Content: "42 is the answer", Score: 0.85},
				}

				// Setup result that will be returned by vector storage
				result := models.SearchResult{
					Answer: "The meaning of life is 42",
					References: []models.Reference{
						{Content: "Life has meaning", Score: 0.95},
						{Content: "42 is the answer", Score: 0.85},
					},
				}

				// Setup repository to return resource IDs for references
				resourceID1 := uuid.New()
				resourceID2 := uuid.New()

				repo.On("GetResourceIDByReference", mock.Anything, refs[0]).Return(resourceID1, nil)
				repo.On("GetResourceIDByReference", mock.Anything, refs[1]).Return(resourceID2, nil)

				// When vector storage's GetAnswerStream is called, it should:
				// 1. Send references to the refsCh
				// 2. Return the result
				vs.On("GetAnswerStream", mock.Anything, "What is the meaning of life?", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						refsCh := args.Get(2).(chan<- []models.Reference)
						// Send references to channel
						refsCh <- refs
					}).
					Return(result, nil)
			},
			expectedResult: models.SearchResult{
				Answer: "The meaning of life is 42",
				References: []models.Reference{
					{Content: "Life has meaning", Score: 0.95},
					{Content: "42 is the answer", Score: 0.85},
				},
			},
			expectedErr: "",
		},
		{
			name:     "VectorStorage returns error",
			question: "Error question",
			setupMocks: func(vs *mockVectorStorage, repo *mockRepository) {
				vs.On("GetAnswerStream", mock.Anything, "Error question", mock.Anything, mock.Anything).
					Return(models.SearchResult{}, errors.New("vector storage error"))
			},
			expectedResult: models.SearchResult{},
			expectedErr:    "Service.GetAnswerStream: vector storage error",
		},
		{
			name:     "Repository returns error",
			question: "Repository error",
			setupMocks: func(vs *mockVectorStorage, repo *mockRepository) {
				// Setup references that will be "returned" by vector storage
				refs := []models.Reference{
					{Content: "Content causing repository error", Score: 0.95},
				}

				// Setup result that will be returned by vector storage
				result := models.SearchResult{
					Answer: "Some answer",
					References: []models.Reference{
						{Content: "Content causing repository error", Score: 0.95},
					},
				}

				// Setup repository to return error for the reference
				repo.On("GetResourceIDByReference", mock.Anything, refs[0]).Return(uuid.Nil, errors.New("repository error"))

				vs.On("GetAnswerStream", mock.Anything, "Repository error", mock.Anything, mock.Anything).
					Run(func(args mock.Arguments) {
						refsCh := args.Get(2).(chan<- []models.Reference)
						// Send references to channel
						refsCh <- refs
					}).
					Return(result, nil)
			},
			expectedResult: models.SearchResult{},
			expectedErr:    "Service.processResult: Service.processReferences: repository error",
		},
		{
			name:          "Context cancellation",
			question:      "Context cancel",
			contextCancel: true,
			setupMocks: func(vs *mockVectorStorage, repo *mockRepository) {
				// Don't setup any expectations since the context should be cancelled before calls
			},
			expectedResult: models.SearchResult{},
			expectedErr:    "context canceled",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks
			mockVS := new(mockVectorStorage)
			mockRepo := new(mockRepository)
			tc.setupMocks(mockVS, mockRepo)

			// Create service with mocks
			service := NewService(mockVS, mockRepo)

			// Create channels
			refsCh := make(chan []models.Reference, 1)
			chunkCh := make(chan []byte, 10)

			// Create context
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// If test case requires context cancellation, cancel it
			if tc.contextCancel {
				cancel()
				// Give a little time for cancellation to propagate
				time.Sleep(10 * time.Millisecond)
			}

			// Call the method
			_, _, _, err := service.GetAnswerStream(ctx, tc.question)

			// Check results
			if tc.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult.Answer, result.Answer)

				// References might have UUIDs added, so we don't directly compare them
				if len(tc.expectedResult.References) > 0 {
					assert.Equal(t, len(tc.expectedResult.References), len(result.References))
					for i, ref := range tc.expectedResult.References {
						assert.Equal(t, ref.Content, result.References[i].Content)
						assert.Equal(t, ref.Score, result.References[i].Score)
					}
				}
			}

			// Verify all expectations were met
			mockVS.AssertExpectations(t)
			mockRepo.AssertExpectations(t)
		})
	}
}
