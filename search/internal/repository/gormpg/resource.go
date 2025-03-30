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

func (r *Repository) GetResources(ctx context.Context) ([]models.Resource, error) {
	const op = "Repository.GetResources"
	slog.DebugContext(ctx, "Fetching resources from database")

	var resources []models.Resource
	err := r.db.WithContext(ctx).Find(&resources).Error
	if err != nil {
		slog.ErrorContext(ctx, "Failed to fetch resources",
			"op", op,
			"error", err)
		return nil, err
	}

	slog.InfoContext(ctx, "Successfully fetched resources",
		"count", len(resources))
	return resources, nil
}

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

func (r *Repository) GetResourceByID(ctx context.Context, resourceID uuid.UUID) (*models.Resource, error) {
	const op = "Repository.GetResourceByID"
	var resource models.Resource

	err := r.db.WithContext(ctx).
		Model(&models.Resource{}).
		First(&resource, "id = ?", resourceID.String()).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		slog.WarnContext(ctx, "Resource not found for reference",
			"op", op,
			"resource_id", resourceID,
		)
		return nil, fmt.Errorf("%s: resource not found for reference content: %s", op, resourceID.String())
	}

	if err != nil {
		slog.ErrorContext(ctx, "Failed to fetch resource ID",
			"op", op,
			"resource_id", resourceID,
			"error", err,
		)
	}

	slog.InfoContext(ctx, "Successfully fetched resource ID",
		"resource_id", resourceID,
	)

	return &resource, nil
}

func (r *Repository) SaveResource(ctx context.Context, resource models.Resource) (*models.Resource, error) {
	const op = "Repository.SaveResource"
	slog.DebugContext(ctx, "Saving resource to database")

	err := r.db.WithContext(ctx).Create(&resource).Error
	if err != nil {
		slog.ErrorContext(ctx, "Failed to save resource",
			"op", op,
			"resource_type", resource.Type,
			"error", err)
		return nil, err
	}

	slog.InfoContext(ctx, "Successfully saved resource",
		"resource_id", resource.ID)
	return &resource, nil
}

func (r *Repository) UpdateResource(ctx context.Context, resource models.Resource) (*models.Resource, error) {
	const op = "Repository.UpdateResource"
	slog.DebugContext(ctx, "Updating resource in database",
		"resource_id", resource.ID)

	err := r.db.WithContext(ctx).Save(&resource).Error
	if err != nil {
		slog.ErrorContext(ctx, "Failed to update resource",
			"op", op,
			"resource_id", resource.ID,
			"error", err)
		return nil, err
	}

	slog.InfoContext(ctx, "Successfully updated resource",
		"resource_id", resource.ID)
	return &resource, nil
}

func (r *Repository) DeleteResource(ctx context.Context, id uuid.UUID) error {
	const op = "Repository.DeleteResource"
	slog.DebugContext(ctx, "Deleting resource from database",
		"resource_id", id)

	err := r.db.WithContext(ctx).Delete(&models.Resource{}, id).Error
	if err != nil {
		slog.ErrorContext(ctx, "Failed to delete resource",
			"op", op,
			"resource_id", id,
			"error", err)
		return err
	}

	slog.InfoContext(ctx, "Successfully deleted resource",
		"resource_id", id)
	return nil
}
