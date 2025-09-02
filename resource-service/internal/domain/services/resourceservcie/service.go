package resourceservcie

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/nzb3/diploma/resource-service/internal/domain/models/resourcemodel"
)

const ResourceTopicName = "resources"

type resourceRepository interface {
	ResourceOwnedByUser(ctx context.Context, resourceID uuid.UUID, userID uuid.UUID) (bool, error)
	GetResources(ctx context.Context, limit int, offset int) ([]resourcemodel.Resource, error)
	GetResourcesByOwnerID(ctx context.Context, ownerID uuid.UUID, limit int, offset int) ([]resourcemodel.Resource, error)
	GetUsersResourceByID(ctx context.Context, resourceID uuid.UUID, ownerID uuid.UUID) (resourcemodel.Resource, error)
	GetResourceByID(ctx context.Context, resourceID uuid.UUID) (resourcemodel.Resource, error)
	SaveResource(ctx context.Context, resource resourcemodel.Resource) (resourcemodel.Resource, error)
	UpdateUsersResource(ctx context.Context, userID uuid.UUID, resource resourcemodel.Resource) (resourcemodel.Resource, error)
	UpdateResourceStatus(ctx context.Context, resourceID uuid.UUID, status resourcemodel.ResourceStatus) (resourcemodel.Resource, error)
	DeleteUsersResource(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error
}

type contentExtractor interface {
	ExtractContent(ctx context.Context, data []byte, dataType string) (string, error)
}

type eventService interface {
	PublishEvent(ctx context.Context, topic string, eventName string, resourceData interface{}) error
}

type Service struct {
	resourceRepo     resourceRepository
	contentExtractor contentExtractor
	eventService     eventService
	// statusChannels maps resource.ID to resourceStatusUpdate channel
	statusChannels sync.Map
}

func NewService(rr resourceRepository, ce contentExtractor, es eventService) *Service {
	slog.Debug("Initializing resource service",
		"repository_type", fmt.Sprintf("%T", rr))
	return &Service{
		resourceRepo:     rr,
		contentExtractor: ce,
		eventService:     es,
	}
}

// SaveUsersResource saves a new resource with the given content and type.
// It also publishes a resource.created event.
func (s *Service) SaveUsersResource(ctx context.Context, userID uuid.UUID, content []byte, resourceType resourcemodel.ResourceType, name, url string) (resourcemodel.Resource, <-chan resourcemodel.ResourceStatusUpdate, error) {
	const op = "Service.SaveUsersResource"

	resourceStatusUpdateCh := make(chan resourcemodel.ResourceStatusUpdate)

	resource := resourcemodel.NewResource(
		resourcemodel.WithOwnerID(userID),
		resourcemodel.WithRawContent(content),
		resourcemodel.WithType(resourceType),
		resourcemodel.WithName(name),
		resourcemodel.WithURL(url),
		resourcemodel.WithStatus(resourcemodel.ResourceStatusProcessing),
	)

	resource, err := s.extractContent(ctx, resource)
	if err != nil {
		return resourcemodel.Resource{}, resourceStatusUpdateCh, fmt.Errorf("%s: %w", op, err)
	}

	resource, err = s.resourceRepo.SaveResource(ctx, resource)
	if err != nil {
		return resourcemodel.Resource{}, resourceStatusUpdateCh, fmt.Errorf("%s: %w", op, err)
	}

	// Register the status channel in sync.Map for indexation processor.
	// Note that this channel will be closed when the resource is deleted.
	s.statusChannels.Store(resource.ID, resourceStatusUpdateCh)

	err = s.eventService.PublishEvent(ctx, ResourceTopicName, "resource.created", map[string]interface{}{
		"resource_id": resource.ID,
		"owner_id":    resource.OwnerID,
		"name":        resource.Name,
		"type":        resource.Type,
		"status":      resource.Status,
		"created_at":  resource.CreatedAt,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to publish resource created event", "error", err)
		return resourcemodel.Resource{}, resourceStatusUpdateCh, err
	}

	return resource, resourceStatusUpdateCh, nil
}

func (s *Service) GetUsersResources(ctx context.Context, userID uuid.UUID, limit, offset int) ([]resourcemodel.Resource, error) {
	const op = "Service.GetUsersResources"
	slog.DebugContext(ctx, "Fetching resources list")

	if limit == 0 {
		limit = 10
	}

	if offset < 0 {
		offset = 0
	}

	resources, err := s.resourceRepo.GetResourcesByOwnerID(ctx, userID, limit, offset)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to retrieve resources",
			"op", op,
			"error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resources, nil
}

func (s *Service) UpdateUsersResource(ctx context.Context, userID uuid.UUID, resourceID uuid.UUID, name *string, content *[]byte) (resourcemodel.Resource, error) {
	const op = "Service.UpdateUsersResource"

	resource, err := s.GetUsersResourceByID(ctx, userID, resourceID)
	if err != nil {
		return resourcemodel.Resource{}, fmt.Errorf("%s: %w", op, err)
	}

	if name != nil {
		resource.Name = *name
	}

	if content != nil {
		resource.RawContent = *content

		resource.ExtractedContent, err = s.contentExtractor.ExtractContent(ctx, resource.RawContent, string(resource.Type))
		if err != nil {
			return resourcemodel.Resource{}, fmt.Errorf("%s: %w", op, err)
		}
	}

	resource, err = s.resourceRepo.UpdateUsersResource(ctx, userID, resource)
	if err != nil {
		return resourcemodel.Resource{}, fmt.Errorf("%s: %w", op, err)
	}

	err = s.eventService.PublishEvent(ctx, ResourceTopicName, "resource.updated", map[string]interface{}{
		"resource_id": resource.ID,
		"owner_id":    resource.OwnerID,
		"name":        resource.Name,
		"type":        resource.Type,
		"status":      resource.Status,
		"updated_at":  resource.UpdatedAt,
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to publish resource updated event", "error", err)
	}

	return resource, nil
}

func (s *Service) DeleteUsersResource(ctx context.Context, userID uuid.UUID, resourceID uuid.UUID) error {
	const op = "Service.DeleteUsersResource"

	resource, err := s.GetUsersResourceByID(ctx, userID, resourceID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = s.resourceRepo.DeleteUsersResource(ctx, userID, resourceID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	err = s.eventService.PublishEvent(ctx, ResourceTopicName, "resource.deleted", map[string]interface{}{
		"resource_id": resourceID,
		"owner_id":    userID,
		"name":        resource.Name,
		"type":        resource.Type,
		"deleted_at":  time.Now(),
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to publish resource deleted event", "error", err)
	}

	return nil
}

func (s *Service) GetUsersResourceByID(ctx context.Context, userID uuid.UUID, resourceID uuid.UUID) (resourcemodel.Resource, error) {
	const op = "Service.GetUsersResourceByID"

	resource, err := s.resourceRepo.GetUsersResourceByID(ctx, userID, resourceID)
	if err != nil {
		return resourcemodel.Resource{}, fmt.Errorf("%s: %w", op, err)
	}

	return resource, nil
}

func (s *Service) extractContent(ctx context.Context, resource resourcemodel.Resource) (resourcemodel.Resource, error) {
	const op = "Service.extractContent"

	content, err := s.contentExtractor.ExtractContent(ctx, resource.RawContent, string(resource.Type))
	if err != nil {
		return resourcemodel.Resource{}, fmt.Errorf("%s: %w", op, err)
	}
	resource.ExtractedContent = content

	return resource, nil
}

func (s *Service) UpdateResourceStatus(
	ctx context.Context,
	resource resourcemodel.Resource,
	status resourcemodel.ResourceStatus,
) (resourcemodel.Resource, error) {
	const op = "Service.UpdateResourceStatus"

	resource.Status = status

	resource, err := s.resourceRepo.UpdateResourceStatus(ctx, resource.ID, status)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update resource",
			"error", err,
		)
		return resourcemodel.Resource{}, fmt.Errorf("%s: %w", op, err)
	}

	err = s.eventService.PublishEvent(ctx, ResourceTopicName, "resource.status_updated", map[string]interface{}{
		"resource_id": resource.ID,
		"owner_id":    resource.OwnerID,
		"old_status":  resource.Status,
		"new_status":  status,
		"updated_at":  time.Now(),
	})
	if err != nil {
		slog.ErrorContext(ctx, "Failed to publish resource status updated event", "error", err)
	}

	return resource, nil
}

// GetResourceStatusChannel retrieves a status channel for a resource ID
func (s *Service) GetResourceStatusChannel(resourceID uuid.UUID) (chan resourcemodel.ResourceStatusUpdate, bool) {
	value, exists := s.statusChannels.Load(resourceID)
	if !exists {
		return nil, false
	}

	ch, ok := value.(chan resourcemodel.ResourceStatusUpdate)
	if !ok {
		s.statusChannels.Delete(resourceID)
		return nil, false
	}

	return ch, true
}

// RemoveResourceStatusChannel removes a status channel from the map
func (s *Service) RemoveResourceStatusChannel(resourceID uuid.UUID) {
	s.statusChannels.Delete(resourceID)
}

// GetResourceByID retrieves a resource by ID (needed for indexation processor)
func (s *Service) GetResourceByID(ctx context.Context, resourceID uuid.UUID) (resourcemodel.Resource, error) {
	const op = "Service.GetResourceByID"

	resource, err := s.resourceRepo.GetResourceByID(ctx, resourceID)
	if err != nil {
		return resourcemodel.Resource{}, fmt.Errorf("%s: %w", op, err)
	}

	return resource, nil
}
