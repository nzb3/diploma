package resourceservcie

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/nzb3/diploma/search/internal/domain/models"
)

// mockResourceRepository is a mock of resourceRepository interface
type mockResourceRepository struct {
	resources   map[uuid.UUID]models.Resource
	getError    error
	saveError   error
	updateError error
	deleteError error
}

func newMockResourceRepository() *mockResourceRepository {
	return &mockResourceRepository{
		resources: make(map[uuid.UUID]models.Resource),
	}
}

func (m *mockResourceRepository) GetResources(ctx context.Context) ([]models.Resource, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	resources := make([]models.Resource, 0, len(m.resources))
	for _, r := range m.resources {
		resources = append(resources, r)
	}
	return resources, nil
}

func (m *mockResourceRepository) GetResourceByID(ctx context.Context, resourceID uuid.UUID) (*models.Resource, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	resource, exists := m.resources[resourceID]
	if !exists {
		return nil, errors.New("resource not found")
	}
	return &resource, nil
}

func (m *mockResourceRepository) SaveResource(ctx context.Context, resource models.Resource) (*models.Resource, error) {
	if m.saveError != nil {
		return nil, m.saveError
	}
	if resource.ID == uuid.Nil {
		resource.ID = uuid.New()
	}
	resource.CreatedAt = time.Now()
	resource.UpdatedAt = time.Now()
	m.resources[resource.ID] = resource
	return &resource, nil
}

func (m *mockResourceRepository) UpdateResource(ctx context.Context, resource models.Resource) (*models.Resource, error) {
	if m.updateError != nil {
		return nil, m.updateError
	}
	existing, exists := m.resources[resource.ID]
	if !exists {
		return nil, errors.New("resource not found")
	}
	resource.CreatedAt = existing.CreatedAt
	resource.UpdatedAt = time.Now()
	m.resources[resource.ID] = resource
	return &resource, nil
}

func (m *mockResourceRepository) DeleteResource(ctx context.Context, id uuid.UUID) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	if _, exists := m.resources[id]; !exists {
		return errors.New("resource not found")
	}
	delete(m.resources, id)
	return nil
}

// mockResourceProcessor is a mock of resourceProcessor interface
type mockResourceProcessor struct {
	processError  error
	extractError  error
	processedData string
	extractedData string
}

func newMockResourceProcessor() *mockResourceProcessor {
	return &mockResourceProcessor{
		processedData: "processed content",
		extractedData: "extracted content",
	}
}

func (m *mockResourceProcessor) ProcessResource(ctx context.Context, resource models.Resource) (models.Resource, error) {
	if m.processError != nil {
		return models.Resource{}, m.processError
	}
	resource.ExtractedContent = m.processedData
	resource.ChunkIDs = []string{uuid.New().String(), uuid.New().String()}
	return resource, nil
}

func (m *mockResourceProcessor) ExtractContent(ctx context.Context, resource models.Resource) (models.Resource, error) {
	if m.extractError != nil {
		return models.Resource{}, m.extractError
	}
	resource.ExtractedContent = m.extractedData
	return resource, nil
}

func createTestResource() models.Resource {
	return models.Resource{
		ID:         uuid.New(),
		Name:       "Test Resource",
		Type:       "document",
		URL:        "https://example.com/test",
		RawContent: []byte("Test content for resource"),
		Status:     models.ResourceStatusProcessing,
	}
}

func TestService_GetResources(t *testing.T) {
	tests := []struct {
		name          string
		setupRepo     func(repo *mockResourceRepository)
		expectedCount int
		expectError   bool
	}{
		{
			name: "successful retrieval",
			setupRepo: func(repo *mockResourceRepository) {
				repo.resources[uuid.New()] = createTestResource()
				repo.resources[uuid.New()] = createTestResource()
			},
			expectedCount: 2,
			expectError:   false,
		},
		{
			name: "empty resources",
			setupRepo: func(repo *mockResourceRepository) {
				// Leave empty
			},
			expectedCount: 0,
			expectError:   false,
		},
		{
			name: "repository error",
			setupRepo: func(repo *mockResourceRepository) {
				repo.getError = errors.New("database connection failed")
			},
			expectedCount: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := newMockResourceRepository()
			procMock := newMockResourceProcessor()
			tt.setupRepo(repoMock)

			service := NewService(repoMock, procMock)
			resources, err := service.GetResources(context.Background())

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(resources))
			}
		})
	}
}

func TestService_GetResourceByID(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(repo *mockResourceRepository) uuid.UUID
		expectError bool
	}{
		{
			name: "resource exists",
			setupRepo: func(repo *mockResourceRepository) uuid.UUID {
				resource := createTestResource()
				repo.resources[resource.ID] = resource
				return resource.ID
			},
			expectError: false,
		},
		{
			name: "resource not found",
			setupRepo: func(repo *mockResourceRepository) uuid.UUID {
				return uuid.New() // ID not in map
			},
			expectError: true,
		},
		{
			name: "repository error",
			setupRepo: func(repo *mockResourceRepository) uuid.UUID {
				repo.getError = errors.New("database connection failed")
				return uuid.New()
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := newMockResourceRepository()
			procMock := newMockResourceProcessor()
			resourceID := tt.setupRepo(repoMock)

			service := NewService(repoMock, procMock)
			resource, err := service.GetResourceByID(context.Background(), resourceID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, resourceID, resource.ID)
			}
		})
	}
}

func TestService_DeleteResource(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(repo *mockResourceRepository) uuid.UUID
		expectError bool
	}{
		{
			name: "successful deletion",
			setupRepo: func(repo *mockResourceRepository) uuid.UUID {
				resource := createTestResource()
				repo.resources[resource.ID] = resource
				return resource.ID
			},
			expectError: false,
		},
		{
			name: "resource not found",
			setupRepo: func(repo *mockResourceRepository) uuid.UUID {
				return uuid.New() // ID not in map
			},
			expectError: true,
		},
		{
			name: "repository error",
			setupRepo: func(repo *mockResourceRepository) uuid.UUID {
				resource := createTestResource()
				repo.resources[resource.ID] = resource
				repo.deleteError = errors.New("deletion failed")
				return resource.ID
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := newMockResourceRepository()
			procMock := newMockResourceProcessor()
			resourceID := tt.setupRepo(repoMock)

			service := NewService(repoMock, procMock)
			err := service.DeleteResource(context.Background(), resourceID)

			if tt.expectError {
				assert.Error(t, err)
				if repoMock.deleteError == nil {
					// Still in map if error was "not found"
					_, exists := repoMock.resources[resourceID]
					assert.False(t, exists)
				}
			} else {
				assert.NoError(t, err)
				_, exists := repoMock.resources[resourceID]
				assert.False(t, exists)
			}
		})
	}
}

func TestService_UpdateResource(t *testing.T) {
	tests := []struct {
		name        string
		setupRepo   func(repo *mockResourceRepository) models.Resource
		expectError bool
	}{
		{
			name: "successful update",
			setupRepo: func(repo *mockResourceRepository) models.Resource {
				resource := createTestResource()
				repo.resources[resource.ID] = resource

				updated := resource
				updated.Name = "Updated Name"
				return updated
			},
			expectError: false,
		},
		{
			name: "resource not found",
			setupRepo: func(repo *mockResourceRepository) models.Resource {
				resource := createTestResource()
				// Not adding to map
				return resource
			},
			expectError: true,
		},
		{
			name: "repository error",
			setupRepo: func(repo *mockResourceRepository) models.Resource {
				resource := createTestResource()
				repo.resources[resource.ID] = resource
				repo.updateError = errors.New("update failed")
				return resource
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := newMockResourceRepository()
			procMock := newMockResourceProcessor()
			resourceToUpdate := tt.setupRepo(repoMock)

			service := NewService(repoMock, procMock)
			updated, err := service.UpdateResource(context.Background(), resourceToUpdate)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, resourceToUpdate.Name, updated.Name)
				assert.Equal(t, resourceToUpdate.ID, updated.ID)
			}
		})
	}
}

func TestService_SaveResource(t *testing.T) {
	tests := []struct {
		name             string
		resource         models.Resource
		setupMocks       func(repo *mockResourceRepository, proc *mockResourceProcessor)
		expectError      bool
		validateResource func(t *testing.T, resource models.Resource)
	}{
		{
			name: "successful save and process",
			resource: models.Resource{
				Name:       "New Resource",
				Type:       "document",
				URL:        "https://example.com/new",
				RawContent: []byte("New resource content"),
			},
			setupMocks: func(repo *mockResourceRepository, proc *mockResourceProcessor) {
				// No errors
			},
			expectError: false,
			validateResource: func(t *testing.T, resource models.Resource) {
				assert.NotEqual(t, uuid.Nil, resource.ID)
				assert.Equal(t, models.ResourceStatusCompleted, resource.Status)
				assert.Equal(t, "processed content", resource.ExtractedContent)
				assert.NotEmpty(t, resource.ChunkIDs)
			},
		},
		{
			name: "save repository error",
			resource: models.Resource{
				Name:       "Error Resource",
				Type:       "document",
				RawContent: []byte("Content that will cause error"),
			},
			setupMocks: func(repo *mockResourceRepository, proc *mockResourceProcessor) {
				repo.saveError = errors.New("save failed")
			},
			expectError: true,
			validateResource: func(t *testing.T, resource models.Resource) {
				// Nothing to validate on error
			},
		},
		{
			name: "process error",
			resource: models.Resource{
				Name:       "Process Error Resource",
				Type:       "document",
				RawContent: []byte("Content that will process error"),
			},
			setupMocks: func(repo *mockResourceRepository, proc *mockResourceProcessor) {
				proc.processError = errors.New("processing failed")
			},
			expectError: true,
			validateResource: func(t *testing.T, resource models.Resource) {
				// Nothing to validate on error
			},
		},
		{
			name: "extract error",
			resource: models.Resource{
				Name:       "Extract Error Resource",
				Type:       "document",
				RawContent: []byte("Content that will extract error"),
			},
			setupMocks: func(repo *mockResourceRepository, proc *mockResourceProcessor) {
				proc.extractError = errors.New("extraction failed")
			},
			expectError: true,
			validateResource: func(t *testing.T, resource models.Resource) {
				// Nothing to validate on error
			},
		},
		{
			name: "empty name sets default",
			resource: models.Resource{
				Name:       "", // Empty name
				Type:       "document",
				RawContent: []byte("This content will be used for the default name"),
			},
			setupMocks: func(repo *mockResourceRepository, proc *mockResourceProcessor) {
				// No errors
			},
			expectError: false,
			validateResource: func(t *testing.T, resource models.Resource) {
				assert.NotEmpty(t, resource.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repoMock := newMockResourceRepository()
			procMock := newMockResourceProcessor()
			tt.setupMocks(repoMock, procMock)

			service := NewService(repoMock, procMock)

			// Create a status update channel
			statusChan := make(chan models.ResourceStatusUpdate, 3)

			resource, err := service.SaveResource(context.Background(), tt.resource, statusChan)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.validateResource(t, resource)

				// Check status channel
				select {
				case update := <-statusChan:
					assert.Equal(t, resource.ID, update.ResourceID)
				default:
					t.Fatal("Expected status update in channel")
				}
			}
		})
	}
}

func TestService_NewService(t *testing.T) {
	repoMock := newMockResourceRepository()
	procMock := newMockResourceProcessor()

	service := NewService(repoMock, procMock)

	assert.NotNil(t, service)
	assert.Same(t, repoMock, service.resourceRepo)
	assert.Same(t, procMock, service.resourceProcessor)
}
