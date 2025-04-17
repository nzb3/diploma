package resourceservcie

import (
	"context"
	"errors"
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
	ExtractContent(ctx context.Context, resource models.Resource) (models.Resource, error)
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

func (s *Service) UpdateResource(ctx context.Context, resource models.Resource) (models.Resource, error) {
	const op = "Service.UpdateResource"
	slog.DebugContext(ctx, "Processing update request",
		"resource_id", resource.ID,
	)
	updatedResource, err := s.resourceRepo.UpdateResource(ctx, resource)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to process resource",
			"op", op,
			"resource_id", resource.ID,
			"error", err,
		)
		return models.Resource{}, fmt.Errorf("%s: %w", op, err)
	}

	slog.InfoContext(ctx, "Successfully updated resource",
		"resource_id", resource.ID,
	)

	return *updatedResource, nil
}

func (s *Service) SaveResource(ctx context.Context, resource models.Resource, statusUpdateCh chan<- models.ResourceStatusUpdate) (models.Resource, error) {
	const op = "Service.SaveResource"

	resource, err := s.runSaveResourcePipeline(ctx, resource, statusUpdateCh)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to save resource",
			"op", op,
			"error", err,
		)
		cleanupErr := s.cleanupProcess(ctx, resource.ID)
		if cleanupErr != nil {
			return models.Resource{}, fmt.Errorf("%s: %w", op, errors.Join(err, cleanupErr))
		}
		return models.Resource{}, fmt.Errorf("%s: %w", op, err)
	}

	slog.InfoContext(ctx, "Successfully saved resource",
		"resource_id", resource.ID,
	)

	return resource, nil
}

func (s *Service) cleanupProcess(ctx context.Context, resourceID uuid.UUID) error {
	const op = "Service.cleanupProcess"

	slog.DebugContext(ctx, "Processing cleanup",
		"resource_id", resourceID,
	)

	if resourceID == uuid.Nil {
		return nil
	}

	err := s.resourceRepo.DeleteResource(ctx, resourceID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete resource",
			"op", op,
			"resource_id", resourceID,
			"error", err,
		)

		return fmt.Errorf("%s: %w", op, err)
	}

	slog.InfoContext(ctx, "Successfully deleted resource",
		"resource_id", resourceID,
	)

	return nil
}

func (s *Service) runSaveResourcePipeline(ctx context.Context, resource models.Resource, statusUpdateCh chan<- models.ResourceStatusUpdate) (models.Resource, error) {
	const op = "Service.runSaveResourcePipeline"

	resource, err := s.saveResource(ctx, resource)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to save resource",
			"op", op,
			"error", err,
		)
		resource.UpdateStatus(models.ResourceStatusFailed, statusUpdateCh)
		return models.Resource{}, fmt.Errorf("%s: %w", op, err)
	}

	resource.UpdateStatus(models.ResourceStatusProcessing, statusUpdateCh)

	resourceID := resource.ID

	resource, err = s.processResource(ctx, resource)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to process resource",
			"op", op,
			"resource_id", resourceID,
			"error", err,
		)
		resource.UpdateStatus(models.ResourceStatusFailed, statusUpdateCh)
		return models.Resource{}, fmt.Errorf("%s: %w", op, err)
	}

	resource.UpdateStatus(models.ResourceStatusCompleted, statusUpdateCh)

	return resource, nil
}

func (s *Service) processResource(ctx context.Context, resource models.Resource) (models.Resource, error) {
	const op = "Service.processResource"
	slog.DebugContext(ctx, "Processing resource",
		"resource_id", resource.ID,
	)

	select {
	case <-ctx.Done():
		return models.Resource{}, ctx.Err()
	default:
		resource, err := s.resourceProcessor.ProcessResource(ctx, resource)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to process resource",
				"op", op,
				"error", err,
			)
			return models.Resource{}, fmt.Errorf("%s: %w", op, err)
		}

		resource, err = s.UpdateResource(ctx, resource)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to update resource",
				"op", op,
				"resource_id", resource.ID,
				"error", err,
			)
			return models.Resource{}, fmt.Errorf("%s: %w", op, err)
		}

		slog.InfoContext(ctx, "Successfully processed resource",
			"resource_id", resource.ID,
		)
		return resource, nil
	}
}

func (s *Service) saveResource(ctx context.Context, resource models.Resource) (models.Resource, error) {
	const op = "Service.saveResource"
	slog.DebugContext(ctx, "Saving resource to repository",
		"resource_type", resource.Type)
	select {
	case <-ctx.Done():
		return models.Resource{}, ctx.Err()
	default:
		slog.Info("Validating resource", "name", resource.Name)
		if resource.Name == "" {
			slog.Info("Setting default resource name")
			resource.SetDefaultName()
			slog.Info("Resource name", "name", resource.Name)
		}

		resource, err := s.resourceProcessor.ExtractContent(ctx, resource)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to save resource",
				"op", op,
				"resource_id", resource.ID,
				"error", err,
			)
			return models.Resource{}, fmt.Errorf("%s: %w", op, err)
		}

		savedResource, err := s.resourceRepo.SaveResource(ctx, resource)
		if err != nil {
			slog.ErrorContext(ctx, "Repository save operation failed",
				"op", op,
				"error", err)
			return models.Resource{}, err
		}

		slog.DebugContext(ctx, "Resource saved successfully",
			"resource_id", savedResource.ID)
		return *savedResource, nil
	}
}
