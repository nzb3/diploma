package resourceservcie

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/nzb3/diploma/resource-service/internal/domain/models/eventmodel"
	"github.com/nzb3/diploma/resource-service/internal/domain/models/resourcemodel"
)

const ResourceTopicName = "resources"

type resourceRepository interface {
	ResourceOwnedByUser(ctx context.Context, resourceID uuid.UUID, userID uuid.UUID) (bool, error)
	GetResources(ctx context.Context, limit int, offset int) ([]resourcemodel.Resource, error)
	GetResourcesByOwnerID(ctx context.Context, ownerID uuid.UUID, limit int, offset int) ([]resourcemodel.Resource, error)
	GetUsersResourceByID(ctx context.Context, resourceID uuid.UUID, ownerID uuid.UUID) (resourcemodel.Resource, error)
	SaveResource(ctx context.Context, resource resourcemodel.Resource) (resourcemodel.Resource, error)
	UpdateUsersResource(ctx context.Context, userID uuid.UUID, resource resourcemodel.Resource) (resourcemodel.Resource, error)
	UpdateResourceStatus(ctx context.Context, resourceID uuid.UUID, status resourcemodel.ResourceStatus) (resourcemodel.Resource, error)
	DeleteUsersResource(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error
}

type contentExtractor interface {
	ExtractContent(ctx context.Context, data []byte, dataType string) (string, error)
}

type eventProvider interface {
	CreateEvent(name string, topic string, payload []byte) error
	ListenEvents(topic string, conditionFn func(eventmodel.Event) bool) (<-chan eventmodel.Event, error)
}

type Service struct {
	resourceRepo     resourceRepository
	contentExtractor contentExtractor
	eventProvider    eventProvider
}

func NewService(rr resourceRepository, ce contentExtractor, ep eventProvider) *Service {
	slog.Debug("Initializing resource service",
		"repository_type", fmt.Sprintf("%T", rr))
	return &Service{
		resourceRepo:     rr,
		contentExtractor: ce,
		eventProvider:    ep,
	}
}

func (s *Service) SaveUsersResource(ctx context.Context, userID uuid.UUID, content []byte, resourceType resourcemodel.ResourceType, name, url string) (<-chan resourcemodel.Resource, <-chan resourcemodel.ResourceStatusUpdate, <-chan error) {
	const op = "Service.SaveUsersResource"

	resourceStatusUpdateCh := make(chan resourcemodel.ResourceStatusUpdate)
	resourceCh := make(chan resourcemodel.Resource)
	errCh := make(chan error)

	resource := resourcemodel.NewResource(
		resourcemodel.WithOwnerID(userID),
		resourcemodel.WithRawContent(content),
		resourcemodel.WithType(resourceType),
		resourcemodel.WithName(name),
		resourcemodel.WithURL(url),
		resourcemodel.WithStatus(resourcemodel.ResourceStatusProcessing),
	)

	go s.saveResource(ctx, resource, resourceCh, resourceStatusUpdateCh, errCh)

	return resourceCh, resourceStatusUpdateCh, errCh
}

func (s *Service) saveResource(
	ctx context.Context,
	resource resourcemodel.Resource,
	resourceCh chan<- resourcemodel.Resource,
	resourceStatusUpdateCh chan<- resourcemodel.ResourceStatusUpdate,
	errCh chan<- error,
) {
	const op = "Service.saveResource"

	defer func() {
		close(resourceCh)
		close(resourceStatusUpdateCh)
		close(errCh)
	}()

	sendErr := func(err error) {
		errCh <- fmt.Errorf("%s: %w", op, err)
	}

	resource, err := s.extractContent(ctx, resource)
	if err != nil {
		sendErr(err)
		return
	}

	resource, err = s.resourceRepo.SaveResource(ctx, resource)
	if err != nil {
		sendErr(err)
		return
	}

	resourceStatusUpdateCh <- resourcemodel.ResourceStatusUpdate{
		ResourceID: resource.ID,
		Status:     resourcemodel.ResourceStatusProcessing,
	}
	resourceCh <- resource

	if err := s.sendResourceCreatedEvent(ctx, resource); err != nil {
		sendErr(err)
		return
	}

	eventCh, err := s.eventProvider.ListenEvents(ResourceTopicName, func(event eventmodel.Event) bool {
		var eventResource resourcemodel.Resource
		if err := json.Unmarshal(event.Payload, &eventResource); err != nil {
			slog.ErrorContext(ctx, "Failed to unmarshal event payload",
				"op", op,
				"error", err,
			)
			return false
		}
		return (event.Name == "resource_indexed" || event.Name == "resource_indexation_failed") && eventResource.ID == resource.ID
	})
	if err != nil {
		sendErr(err)
		return
	}

	select {
	case <-ctx.Done():
		slog.InfoContext(ctx, "Context cancelled, stopping resource processing",
			"op", op,
			"resource_id", resource.ID,
		)
		sendErr(fmt.Errorf("context cancelled"))
	case event := <-eventCh:
		s.handleResourceEvent(ctx, event, resource, resourceStatusUpdateCh, errCh)
	}
}

func (s *Service) handleResourceEvent(
	ctx context.Context,
	event eventmodel.Event,
	resource resourcemodel.Resource,
	resourceStatusUpdateCh chan<- resourcemodel.ResourceStatusUpdate,
	errCh chan<- error,
) {
	const op = "Service.handleResourceEvent"
	var err error

	switch event.Name {
	case "resource_indexed":
		resource, err = s.UpdateResourceStatus(ctx, resource, resourcemodel.ResourceStatusCompleted)
		if err != nil {
			errCh <- fmt.Errorf("%s: %w", op, err)
			return
		}
		resourceStatusUpdateCh <- resourcemodel.ResourceStatusUpdate{
			ResourceID: resource.ID,
			Status:     resourcemodel.ResourceStatusCompleted,
		}
	case "resource_indexation_failed":
		resource, err = s.UpdateResourceStatus(ctx, resource, resourcemodel.ResourceStatusFailed)
		if err != nil {
			errCh <- fmt.Errorf("%s: %w", op, err)
			return
		}
		resourceStatusUpdateCh <- resourcemodel.ResourceStatusUpdate{
			ResourceID: resource.ID,
			Status:     resourcemodel.ResourceStatusFailed,
		}
	}
}

type ResourceCreatedEvent struct {
	ID      uuid.UUID `json:"resource_id"`
	Content string    `json:"resource_content"`
}

func (s *Service) sendResourceCreatedEvent(ctx context.Context, resource resourcemodel.Resource) error {
	const op = "Service.sendResourceCreatedEvent"
	const resourceCreatedEventName = "resource_created"

	data := &ResourceCreatedEvent{
		ID:      resource.ID,
		Content: resource.ExtractedContent,
	}

	event, err := eventmodel.NewEvent(resourceCreatedEventName, ResourceTopicName, data)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create resource created event",
			"op", op,
			"error", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	return s.eventProvider.CreateEvent(event.Name, event.Topic, event.Payload)
}

//resource_created event {
//    resource_id uuid.UUID,
//    resource_content string
//}
//resource_indexed event {
//    resource_id
//}
//resource_indexation_failed event {
//    resource_id
// }

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

	return resource, nil
}

func (s *Service) DeleteUsersResource(ctx context.Context, userID uuid.UUID, resourceID uuid.UUID) error {
	const op = "Service.DeleteUsersResource"

	err := s.resourceRepo.DeleteUsersResource(ctx, userID, resourceID)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
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

	return resource, nil
}
