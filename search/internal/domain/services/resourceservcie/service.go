package resourceservcie

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/nzb3/diploma/search/internal/domain/models"
)

type resourceRepository interface {
	GetResources(ctx context.Context) ([]models.Resource, error)
	GetResourceByID(ctx context.Context, resourceID uuid.UUID) (*models.Resource, error)
	SaveResource(ctx context.Context, resource models.Resource) (*models.Resource, error)
	UpdateResource(ctx context.Context, resource models.Resource) (*models.Resource, error)
	DeleteResource(ctx context.Context, id uuid.UUID) error
}

type resourceProcessor interface {
	ProcessResource(ctx context.Context, resource models.Resource) (models.Resource, error)
}

type Service struct {
	resourceRepo      resourceRepository
	resourceProcessor resourceProcessor
}

func NewService(rr resourceRepository, rp resourceProcessor) *Service {
	slog.Debug("Initializing resource service",
		"repository_type", fmt.Sprintf("%T", rr))
	return &Service{
		resourceRepo:      rr,
		resourceProcessor: rp,
	}
}

func (s *Service) GetResources(ctx context.Context) ([]models.Resource, error) {
	const op = "Service.GetResources"
	slog.DebugContext(ctx, "Fetching resources list")

	resources, err := s.resourceRepo.GetResources(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to retrieve resources",
			"op", op,
			"error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	slog.InfoContext(ctx, "Successfully fetched resources",
		"count", len(resources))
	return resources, nil
}

func (s *Service) DeleteResource(ctx context.Context, id uuid.UUID) error {
	const op = "Service.DeleteResource"
	slog.DebugContext(ctx, "Processing delete request",
		"resource_id", id,
	)

	err := s.resourceRepo.DeleteResource(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "Resource deletion failed",
			"op", op,
			"resource_id", id,
			"error", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	slog.InfoContext(ctx, "Resource deleted successfully",
		"resource_id", id)
	return nil
}

func (s *Service) GetResourceByID(ctx context.Context, resourceID uuid.UUID) (models.Resource, error) {
	const op = "Service.GetResourceByID"
	slog.DebugContext(ctx, "Processing get request",
		"resource_id", resourceID,
	)

	resource, err := s.resourceRepo.GetResourceByID(ctx, resourceID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to retrieve resource",
			"op", op,
			"resource_id", resourceID,
			"error", err,
		)
		return models.Resource{}, fmt.Errorf("%s: %w", op, err)
	}

	slog.InfoContext(ctx, "Successfully fetched resource",
		"resource_id", resourceID,
	)
	return *resource, nil
}

func (s *Service) SaveResource(ctx context.Context, resource models.Resource) (<-chan models.Resource, <-chan error) {
	const op = "Service.SaveResource"
	resourceChan := make(chan models.Resource)
	errChan := make(chan error)

	go func() {
		defer func() {
			close(resourceChan)
			close(errChan)
			slog.DebugContext(ctx, "Resource channels closed")
		}()

		slog.InfoContext(ctx, "Starting resource processing",
			"resource_type", resource.Type,
			"content_size", len(resource.RawContent))

		savedResource, err := s.saveResource(ctx, resource)
		if err != nil {
			slog.ErrorContext(ctx, "Initial save failed",
				"op", op,
				"error", err)
			errChan <- fmt.Errorf("%s: %w", op, err)
			return
		}

		resourceChan <- *savedResource
		slog.DebugContext(ctx, "Initial resource state saved",
			"resource_id", savedResource.ID)

		processedResource, err := s.resourceProcessor.ProcessResource(ctx, *savedResource)
		if err != nil {
			slog.ErrorContext(ctx, "Resource processing failed",
				"op", op,
				"resource_id", savedResource.ID,
				"error", err)
			errChan <- fmt.Errorf("%s: %w", op, err)
			return
		}

		slog.DebugContext(ctx, "Updating resource status", "resource_id", resource.ID)
		updatedResource, err := s.resourceRepo.UpdateResource(ctx, processedResource)
		if err != nil {
			slog.ErrorContext(ctx, "Resource update failed",
				"op", op,
				"resource_id", resource.ID,
				"error", err)
			errChan <- err
			return
		}

		slog.DebugContext(ctx, "Resource status updated",
			"resource_id", updatedResource.ID,
			"new_status", updatedResource.Status)

		resourceChan <- *updatedResource
		slog.InfoContext(ctx, "Resource processing completed",
			"resource_id", updatedResource.ID,
			"status", updatedResource.Status)
	}()

	return resourceChan, errChan
}

func (s *Service) saveResource(ctx context.Context, resource models.Resource) (*models.Resource, error) {
	const op = "Service.saveResource"
	slog.DebugContext(ctx, "Saving resource to repository",
		"resource_type", resource.Type)

	slog.Info("Validating resource", "name", resource.Name)
	if resource.Name == "" {
		slog.Info("Setting default resource name")
		resource.SetDefaultName()
		slog.Info("Resource name", "name", resource.Name)
	}

	savedResource, err := s.resourceRepo.SaveResource(ctx, resource)
	if err != nil {
		slog.ErrorContext(ctx, "Repository save operation failed",
			"op", op,
			"error", err)
		return nil, err
	}

	slog.DebugContext(ctx, "Resource saved successfully",
		"resource_id", savedResource.ID)
	return savedResource, nil
}
