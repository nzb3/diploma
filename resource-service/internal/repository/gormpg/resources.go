package gormpg

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"github.com/nzb3/diploma/resource-service/internal/domain/models/resourcemodel"
	"github.com/nzb3/diploma/resource-service/internal/repository/gormpg/dto"
)

// GetResources retrieves all resources from the database
func (r *Repository) GetResources(ctx context.Context, ownerID uuid.UUID, limit int, offset int) ([]resourcemodel.Resource, error) {
	const op = "Repository.GetUsersResources"

	var resources []dto.Resource
	if err := r.db.WithContext(ctx).Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return lo.Map(resources, func(resource dto.Resource, _ int) resourcemodel.Resource {
		return *resource.ToDomain()
	}), nil
}

// GetResourcesByOwnerID retrieves all resources belonging to a specific owner
func (r *Repository) GetResourcesByOwnerID(ctx context.Context, ownerID uuid.UUID, limit int, offset int) ([]resourcemodel.Resource, error) {
	const op = "Repository.GetResourcesByOwnerID"

	var resources []dto.Resource
	if err := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return lo.Map(resources, func(resource dto.Resource, _ int) resourcemodel.Resource {
		return *resource.ToDomain()
	}), nil
}

// GetResourceByID retrieves a resource by its ID
func (r *Repository) GetResourceByID(ctx context.Context, resourceID uuid.UUID) (*resourcemodel.Resource, error) {
	const op = "Repository.GetUsersResourceByID"

	var resource dto.Resource
	if err := r.db.WithContext(ctx).Where("id = ?", resourceID).First(&resource).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return resource.ToDomain(), nil
}

// SaveResource saves a new resource to the database
func (r *Repository) SaveResource(ctx context.Context, resource resourcemodel.Resource) (*resourcemodel.Resource, error) {
	const op = "Repository.SaveUsersResource"
	resourceEntity := dto.Resource{}
	resourceEntity.FeelFromDomain(&resource)

	if err := r.db.WithContext(ctx).Create(&resourceEntity).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return resourceEntity.ToDomain(), nil
}

// UpdateResource updates an existing resource in the database
func (r *Repository) UpdateResource(ctx context.Context, resource resourcemodel.Resource) (*resourcemodel.Resource, error) {
	const op = "Repository.updateResource"

	updates := map[string]interface{}{}

	if resource.Name != "" {
		updates["name"] = resource.Name
	}
	if resource.Type != "" {
		updates["type"] = resource.Type
	}
	if resource.URL != "" {
		updates["url"] = resource.URL
	}
	if resource.ExtractedContent != "" {
		updates["extracted_content"] = resource.ExtractedContent
	}
	if len(resource.RawContent) > 0 {
		updates["raw_content"] = resource.RawContent
	}
	if resource.Status != "" {
		updates["status"] = resource.Status
	}
	if resource.OwnerID != uuid.Nil {
		updates["owner_id"] = resource.OwnerID
	}

	if len(updates) == 0 {
		var currentResource dto.Resource
		if err := r.db.WithContext(ctx).First(&currentResource, resource.ID).Error; err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		return currentResource.ToDomain(), nil
	}

	resourceToUpdate := dto.Resource{
		ID: resource.ID,
	}

	if err := r.db.WithContext(ctx).Model(&resourceToUpdate).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	var updatedResource dto.Resource
	if err := r.db.WithContext(ctx).First(&updatedResource, resource.ID).Error; err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return updatedResource.ToDomain(), nil
}

// DeleteResource deletes a resource by its ID
func (r *Repository) DeleteResource(ctx context.Context, id uuid.UUID) error {
	const op = "Repository.DeleteUsersResource"

	var resource dto.Resource
	resource.ID = id

	if err := r.db.WithContext(ctx).Delete(&resource).Error; err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *Repository) ResourceOwnedByUser(ctx context.Context, resourceID uuid.UUID, userID uuid.UUID) (bool, error) {
	const op = "Repository.ResourceOwnedByUser"

	var count int64
	err := r.db.WithContext(ctx).
		Model(&dto.Resource{}).
		Where("id = ? AND (owner_id = ? OR owner_id IS NULL OR owner_id = '')", resourceID, userID).
		Count(&count).
		Error
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return count > 0, nil
}
