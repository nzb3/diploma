package resourceservcie

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/nzb3/diploma/search/internal/controllers/middleware"
	"github.com/nzb3/diploma/search/internal/domain/models"
)

type resourceRepository interface {
	ResourceOwnedByUser(ctx context.Context, resourceID uuid.UUID, userID string) (bool, error)
	GetResources(ctx context.Context) ([]models.Resource, error)
	GetResourcesByOwnerID(ctx context.Context, ownerID string) ([]models.Resource, error)
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

// TODO: move to context provider
// getUserID attempts to get the authenticated user ID from context
// If not found, returns an error
func getUserID(ctx context.Context) (string, error) {
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		return "", errors.New("user not authenticated")
	}
	return userID, nil
}

func (s *Service) GetResources(ctx context.Context) ([]models.Resource, error) {
	const op = "Service.GetResources"
	slog.DebugContext(ctx, "Fetching resources list")

	userID, err := getUserID(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get user ID",
			"op", op,
			"error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	slog.InfoContext(ctx, "Fetching resources for user", "user_id", userID)

	resources, err := s.resourceRepo.GetResourcesByOwnerID(ctx, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to retrieve resources",
			"op", op,
			"error", err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return resources, nil
}

func (s *Service) DeleteResource(ctx context.Context, id uuid.UUID) error {
	const op = "Service.DeleteResource"

	userID, err := getUserID(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get user ID",
			"op", op,
			"error", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	slog.DebugContext(ctx, "Processing delete request",
		"resource_id", id,
		"user_id", userID,
	)

	err = s.checkOwnership(ctx, id, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete resource",
			"op", op,
			"error", err,
		)
		return fmt.Errorf("%s: %w", op, err)
	}

	err = s.resourceRepo.DeleteResource(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "Resource deletion failed",
			"op", op,
			"resource_id", id,
			"error", err)
		return fmt.Errorf("%s: %w", op, err)
	}

	slog.InfoContext(ctx, "Resource deleted successfully",
		"resource_id", id,
		"user_id", userID)
	return nil
}

func (s *Service) GetResourceByID(ctx context.Context, resourceID uuid.UUID) (models.Resource, error) {
	const op = "Service.GetResourceByID"

	userID, err := getUserID(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get user ID",
			"op", op,
			"error", err)
		return models.Resource{}, fmt.Errorf("%s: %w", op, err)
	}

	err = s.checkOwnership(ctx, resourceID, userID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete resource",
			"op", op,
			"resource_id", resourceID,
			"user_id", userID,
			"error", err,
		)
		return models.Resource{}, fmt.Errorf("%s: %w", op, err)
	}

	slog.DebugContext(ctx, "Processing get request",
		"resource_id", resourceID,
		"user_id", userID,
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
		"user_id", userID,
	)
	return *resource, nil
}

func (s *Service) UpdateResource(ctx context.Context, resource models.Resource) (models.Resource, error) {
	const op = "Service.UpdateResource"
	slog.DebugContext(ctx, "Processing update request",
		"resource_id", resource.ID,
	)

	err := s.checkOwnership(ctx, resource.ID, resource.OwnerID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update resource",
			"op", op,
			"resource_id", resource.ID,
			"user_id", resource.OwnerID,
			"error", err,
		)
		return models.Resource{}, fmt.Errorf("%s: %w", op, err)
	}

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

	userID, err := getUserID(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get user ID",
			"op", op,
			"error", err)
		return models.Resource{}, fmt.Errorf("%s: %w", op, err)
	}

	resource.OwnerID = userID

	slog.InfoContext(ctx, "Saving resource for user",
		"user_id", userID,
		"resource_type", resource.Type)

	resource, err = s.runSaveResourcePipeline(ctx, resource, statusUpdateCh)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to save resource",
			"op", op,
			"error", err,
			"user_id", userID,
		)

		return models.Resource{}, fmt.Errorf("%s: %w", op, err)
	}

	slog.InfoContext(ctx, "Successfully saved resource",
		"resource_id", resource.ID,
		"user_id", userID,
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

	resource, err := s.saveResource(ctx, resource, statusUpdateCh)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to save resource",
			"op", op,
			"error", err,
		)

		_, updateStatusErr := s.UpdateResourceStatus(ctx, resource, models.ResourceStatusFailed, statusUpdateCh)
		if updateStatusErr != nil {
			return models.Resource{}, fmt.Errorf("%s: %w", op, errors.Join(err, updateStatusErr))
		}

		return models.Resource{}, fmt.Errorf("%s: %w", op, err)
	}

	ctx, cancel := context.WithTimeout(
		context.WithValue(
			context.Background(),
			"user_id",
			resource.OwnerID,
		), 20*time.Minute)
	go func() {
		defer cancel()

		resourceID := resource.ID

		resource, err = s.processResource(ctx, resource, statusUpdateCh)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to process resource",
				"op", op,
				"resource_id", resourceID,
				"error", err,
			)

			_, updateStatusErr := s.UpdateResourceStatus(ctx, resource, models.ResourceStatusFailed, statusUpdateCh)
			if updateStatusErr != nil {
				slog.ErrorContext(ctx, "Failed to update resource",
					"op", op,
					"resource_id", resourceID,
					"error", err,
				)
				return
			}
			return
		}
	}()

	return resource, nil
}

func (s *Service) processResource(ctx context.Context, resource models.Resource, statusUpdateCh chan<- models.ResourceStatusUpdate) (models.Resource, error) {
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

		resource, err = s.UpdateResourceStatus(ctx, resource, models.ResourceStatusCompleted, statusUpdateCh)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to update resource",
				"op", op,
				"resource_id", resource.ID,
				"error", err,
			)
		}

		return resource, nil
	}
}

func (s *Service) saveResource(ctx context.Context, resource models.Resource, statusUpdateCh chan<- models.ResourceStatusUpdate) (models.Resource, error) {
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

		resource, err = s.UpdateResourceStatus(ctx, *savedResource, models.ResourceStatusProcessing, statusUpdateCh)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to update resource",
				"op", op,
				"resource_id", resource.ID,
				"error", err,
			)
		}

		slog.DebugContext(ctx, "Resource saved successfully",
			"resource_id", savedResource.ID)
		return resource, nil
	}
}

func (s *Service) UpdateResourceStatus(
	ctx context.Context,
	resource models.Resource,
	status models.ResourceStatus,
	updateCh ...chan<- models.ResourceStatusUpdate,
) (models.Resource, error) {
	const op = "Service.UpdateResourceStatus"

	resource.Status = status

	resource, err := s.UpdateResource(ctx, resource)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update resource",
			"error", err,
		)
		return models.Resource{}, fmt.Errorf("%s: %w", op, err)
	}

	if len(updateCh) > 0 {
		for _, u := range updateCh {
			u <- models.ResourceStatusUpdate{
				ResourceID: resource.ID,
				Status:     resource.Status,
			}
		}
	}

	return resource, nil
}

func (s *Service) checkOwnership(ctx context.Context, resourceID uuid.UUID, userID string) error {
	const op = "Service.checkOwnership"
	owned, err := s.resourceRepo.ResourceOwnedByUser(ctx, resourceID, userID)
	if err != nil {
		slog.Error("Failed to check ownership of resource",
			"op", op,
			"resource_id", resourceID,
			"user_id", userID,
			"error", err,
		)
		return err
	}

	if !owned {
		return errors.New("user haven't owned resource")
	}

	return nil
}
