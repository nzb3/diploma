package resources

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/lo"

	"github.com/nzb3/diploma/resource-service/database/sqlc"
	"github.com/nzb3/diploma/resource-service/internal/domain/models/resourcemodel"
	"github.com/nzb3/diploma/resource-service/internal/repository/pgx"
)

type baseRepository interface {
	Close()
	DB() *pgxpool.Pool
	Queries() *sqlc.Queries
	Health(ctx context.Context) error
}

type Repository struct {
	baseRepository
}

func NewResourceRepository(ctx context.Context, repository baseRepository) *Repository {
	return &Repository{
		baseRepository: repository,
	}
}

// ResourceOwnedByUser checks if a resource is owned by a specific user
func (r *Repository) ResourceOwnedByUser(ctx context.Context, resourceID uuid.UUID, userID uuid.UUID) (bool, error) {
	params := sqlc.CheckResourceOwnershipParams{
		ID:      pgx.UuidToPgType(resourceID),
		OwnerID: pgx.UuidToPgType(userID),
	}

	owned, err := r.Queries().CheckResourceOwnership(ctx, params)
	if err != nil {
		return false, fmt.Errorf("failed to check resource ownership: %w", err)
	}

	return owned, nil
}

// GetResources retrieves all resources
func (r *Repository) GetResources(ctx context.Context, limit int, offset int) ([]resourcemodel.Resource, error) {
	sqlcResources, err := r.Queries().GetResources(ctx, sqlc.GetResourcesParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get resources: %w", err)
	}

	return lo.Map(sqlcResources, func(sqlcResource sqlc.Resources, _ int) resourcemodel.Resource {
		return sqlcResourceToModel(sqlcResource)
	}), nil
}

// GetResourcesByOwnerID retrieves all resources by owner ID
func (r *Repository) GetResourcesByOwnerID(ctx context.Context, ownerID uuid.UUID, limit int, offset int) ([]resourcemodel.Resource, error) {
	sqlcResources, err := r.Queries().GetResourcesByOwnerID(ctx, sqlc.GetResourcesByOwnerIDParams{
		OwnerID: pgx.UuidToPgType(ownerID),
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get resources by owner id: %w", err)
	}

	return lo.Map(sqlcResources, func(sqlcResource sqlc.Resources, _ int) resourcemodel.Resource {
		return sqlcResourceToModel(sqlcResource)
	}), nil
}

// GetResourceByID retrieves a resource by ID
func (r *Repository) GetUsersResourceByID(ctx context.Context, resourceID uuid.UUID, ownerID uuid.UUID) (resourcemodel.Resource, error) {
	sqlcResource, err := r.Queries().GetUsersResourceByID(ctx, sqlc.GetUsersResourceByIDParams{
		ID:      pgx.UuidToPgType(resourceID),
		OwnerID: pgx.UuidToPgType(ownerID),
	})
	if err != nil {
		return resourcemodel.Resource{}, fmt.Errorf("failed to get resource by ID: %w", err)
	}

	resource := sqlcResourceToModel(sqlcResource)
	return resource, nil
}

// SaveResource creates a new resource
func (r *Repository) SaveResource(ctx context.Context, resource resourcemodel.Resource) (resourcemodel.Resource, error) {
	params := sqlc.CreateResourceParams{
		Name:             resource.Name,
		Type:             modelTypeToSqlc(resource.Type),
		Url:              pgx.StringToPgType(resource.URL),
		ExtractedContent: pgx.StringToPgType(resource.ExtractedContent),
		RawContent:       resource.RawContent,
		OwnerID:          pgx.UuidToPgType(resource.OwnerID),
	}

	sqlcResource, err := r.Queries().CreateResource(ctx, params)
	if err != nil {
		return resourcemodel.Resource{}, fmt.Errorf("failed to save resource: %w", err)
	}

	savedResource := sqlcResourceToModel(sqlcResource)
	return savedResource, nil
}

// UpdateUsersResource updates an existing resource
func (r *Repository) UpdateUsersResource(ctx context.Context, userID uuid.UUID, resource resourcemodel.Resource) (resourcemodel.Resource, error) {
	if userID != resource.OwnerID {
		return resourcemodel.Resource{}, fmt.Errorf("user id must be owned by another user")
	}

	params := sqlc.UpdateUsersResourceParams{
		ID:               pgx.UuidToPgType(resource.ID),
		Name:             resource.Name,
		Type:             sqlc.ResourceType(resource.Type),
		Url:              pgx.StringToPgType(resource.URL),
		ExtractedContent: pgx.StringToPgType(resource.ExtractedContent),
		RawContent:       resource.RawContent,
		Status:           sqlc.ResourceStatus(resource.Status),
		OwnerID:          pgx.UuidToPgType(userID),
	}

	sqlcResource, err := r.Queries().UpdateUsersResource(ctx, params)
	if err != nil {
		return resourcemodel.Resource{}, fmt.Errorf("failed to update resource: %w", err)
	}

	updatedResource := sqlcResourceToModel(sqlcResource)
	return updatedResource, nil
}

// UpdateResourceStatus update status of resource
func (r *Repository) UpdateResourceStatus(ctx context.Context, resourceID uuid.UUID, status resourcemodel.ResourceStatus) (resourcemodel.Resource, error) {
	sqlcResource, err := r.Queries().UpdateResourceStatus(ctx, sqlc.UpdateResourceStatusParams{
		ID:     pgx.UuidToPgType(resourceID),
		Status: sqlc.ResourceStatus(status),
	})
	if err != nil {
		return resourcemodel.Resource{}, fmt.Errorf("failed to update resource status: %w", err)
	}

	updatedResource := sqlcResourceToModel(sqlcResource)
	return updatedResource, nil
}

// DeleteUsersResource deletes a resource by ID
func (r *Repository) DeleteUsersResource(ctx context.Context, id uuid.UUID, ownerID uuid.UUID) error {
	err := r.Queries().DeleteUsersResource(ctx, sqlc.DeleteUsersResourceParams{
		ID:      pgx.UuidToPgType(id),
		OwnerID: pgx.UuidToPgType(ownerID),
	})
	if err != nil {
		return fmt.Errorf("failed to delete resource: %w", err)
	}

	return nil
}

func modelTypeToSqlc(resourceType resourcemodel.ResourceType) sqlc.ResourceType {
	switch resourceType {
	case resourcemodel.ResourceTypePDF:
		return sqlc.ResourceTypePdf
	case resourcemodel.ResourceTypeText:
		return sqlc.ResourceTypeTxt
	case resourcemodel.ResourceTypeURL:
		return sqlc.ResourceTypeUrl
	default:
		return sqlc.ResourceTypeTxt
	}
}

func sqlcTypeToModel(resourceType sqlc.ResourceType) resourcemodel.ResourceType {
	switch resourceType {
	case sqlc.ResourceTypePdf:
		return resourcemodel.ResourceTypePDF
	case sqlc.ResourceTypeTxt:
		return resourcemodel.ResourceTypeText
	case sqlc.ResourceTypeUrl:
		return resourcemodel.ResourceTypeURL
	default:
		return resourcemodel.ResourceTypeText
	}
}

func modelStatusToSqlc(status resourcemodel.ResourceStatus) sqlc.ResourceStatus {
	switch status {
	case resourcemodel.ResourceStatusPending:
		return sqlc.ResourceStatusPending
	case resourcemodel.ResourceStatusProcessing:
		return sqlc.ResourceStatusProcessing
	case resourcemodel.ResourceStatusCompleted:
		return sqlc.ResourceStatusCompleted
	case resourcemodel.ResourceStatusFailed:
		return sqlc.ResourceStatusFailed
	default:
		return sqlc.ResourceStatusPending
	}
}

func sqlcStatusToModel(status sqlc.ResourceStatus) resourcemodel.ResourceStatus {
	switch status {
	case sqlc.ResourceStatusPending:
		return resourcemodel.ResourceStatusPending
	case sqlc.ResourceStatusProcessing:
		return resourcemodel.ResourceStatusProcessing
	case sqlc.ResourceStatusCompleted:
		return resourcemodel.ResourceStatusCompleted
	case sqlc.ResourceStatusFailed:
		return resourcemodel.ResourceStatusFailed
	default:
		return resourcemodel.ResourceStatusPending
	}
}

func sqlcResourceToModel(sqlcResource sqlc.Resources) resourcemodel.Resource {
	return resourcemodel.Resource{
		ID:               pgx.PgTypeToUUID(sqlcResource.ID),
		Name:             sqlcResource.Name,
		Type:             sqlcTypeToModel(sqlcResource.Type),
		URL:              pgx.PgTypeToString(sqlcResource.Url),
		ExtractedContent: pgx.PgTypeToString(sqlcResource.ExtractedContent),
		RawContent:       sqlcResource.RawContent,
		Status:           sqlcStatusToModel(sqlcResource.Status),
		OwnerID:          pgx.PgTypeToUUID(sqlcResource.OwnerID),
		CreatedAt:        sqlcResource.CreatedAt.Time,
		UpdatedAt:        sqlcResource.UpdatedAt.Time,
	}
}
