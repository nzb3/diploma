package gormpg

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/nzb3/diploma/search/internal/domain/models"
)

// GetResources retrieves all resources from the database
func (r *Repository) GetResources(ctx context.Context) ([]models.Resource, error) {
	const op = "Repository.GetResources"

	var resources []models.Resource
	if err := r.db.WithContext(ctx).Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return resources, nil
}

// GetResourceIDByReference retrieves resource by reference
func (r *Repository) GetResourceIDByReference(ctx context.Context, reference models.Reference) (uuid.UUID, error) {
	const op = "Repository.GetResourceIDByReference"
	slog.DebugContext(ctx, "Fetching resource ID by reference",
		"reference_content_length", len(reference.Content))

	var idStr string
	err := r.db.WithContext(ctx).
		Model(&models.Resource{}).
		Select("resources.id").
		Joins("JOIN resource_embedding ON resources.id = resource_embedding.resource_id").
		Joins("JOIN embeddings ON resource_embedding.embedding_id = embeddings.uuid").
		Where("embeddings.document = ?", reference.Content).
		Limit(1).
		Pluck("resources.id", &idStr).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		slog.WarnContext(ctx, "Resource not found for reference",
			"op", op,
			"reference_content", reference.Content)
		return uuid.Nil, fmt.Errorf("%s: resource not found for reference content: %s", op, reference.Content)
	}

	if err != nil {
		slog.ErrorContext(ctx, "Failed to fetch resource ID",
			"op", op,
			"error", err)
		return uuid.Nil, fmt.Errorf("%s: failed to fetch resource ID: %w", op, err)
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to parse resource ID",
			"op", op,
			"id_str", idStr,
			"error", err)
		return uuid.Nil, fmt.Errorf("%s: invalid UUID format: %w", op, err)
	}

	slog.InfoContext(ctx, "Successfully fetched resource ID",
		"resource_id", id)
	return id, nil
}

// GetResourcesByOwnerID retrieves all resources belonging to a specific owner
func (r *Repository) GetResourcesByOwnerID(ctx context.Context, ownerID string) ([]models.Resource, error) {
	const op = "Repository.GetResourcesByOwnerID"

	var resources []models.Resource
	if err := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return resources, nil
}

// GetResourceByID retrieves a resource by its ID
func (r *Repository) GetResourceByID(ctx context.Context, resourceID uuid.UUID) (*models.Resource, error) {
	const op = "Repository.GetResourceByID"

	var resource models.Resource
	if err := r.db.WithContext(ctx).Where("id = ?", resourceID).First(&resource).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &resource, nil
}

// SaveResource saves a new resource to the database
func (r *Repository) SaveResource(ctx context.Context, resource models.Resource) (*models.Resource, error) {
	const op = "Repository.SaveResource"

	if err := r.db.WithContext(ctx).Create(&resource).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &resource, nil
}

// UpdateResource updates an existing resource in the database
func (r *Repository) UpdateResource(ctx context.Context, resource models.Resource) (*models.Resource, error) {
	const op = "Repository.UpdateResource"

	if err := r.db.WithContext(ctx).Save(&resource).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &resource, nil
}

// DeleteResource deletes a resource by its ID
func (r *Repository) DeleteResource(ctx context.Context, id uuid.UUID) error {
	const op = "Repository.DeleteResource"

	var resource models.Resource
	resource.ID = id

	if err := r.db.WithContext(ctx).Delete(&resource).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *Repository) ResourceOwnedByUser(ctx context.Context, resourceID uuid.UUID, userID string) (bool, error) {
	const op = "Repository.ResourceOwnedByUser"

	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.Resource{}).
		Where("id = ? AND (owner_id = ? OR owner_id IS NULL OR owner_id = '')", resourceID, userID).
		Count(&count).
		Error

	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return count > 0, nil
}
