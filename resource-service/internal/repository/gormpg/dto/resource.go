package dto

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/nzb3/diploma/resource-service/internal/domain/models/resourcemodel"
)

type Resource struct {
	ID               uuid.UUID                    `gorm:"type:uuid;primaryKey"`
	Name             string                       `gorm:"type:varchar(255)"`
	Type             resourcemodel.ResourceType   `gorm:"type:varchar(100)"`
	URL              string                       `gorm:"type:varchar(255)"`
	ExtractedContent string                       `gorm:"type:text"`
	RawContent       []byte                       `gorm:"type:bytea"`
	Status           resourcemodel.ResourceStatus `gorm:"type:varchar(50)"`
	OwnerID          uuid.UUID                    `gorm:"type:varchar(100)"`
	CreatedAt        time.Time                    `gorm:"autoCreateTime"`
	UpdatedAt        time.Time                    `gorm:"autoUpdateTime"`
}

func (r *Resource) TableName() string {
	return "resources"
}

func (r *Resource) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}

	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()

	return nil
}

func (r *Resource) BeforeUpdate(tx *gorm.DB) error {
	r.UpdatedAt = time.Now()
	return nil
}

func (r *Resource) ToDomain() *resourcemodel.Resource {
	return &resourcemodel.Resource{
		ID:               r.ID,
		Name:             r.Name,
		Type:             r.Type,
		URL:              r.URL,
		ExtractedContent: r.ExtractedContent,
		RawContent:       r.RawContent,
		Status:           r.Status,
		OwnerID:          r.OwnerID,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}

func (r *Resource) FeelFromDomain(d *resourcemodel.Resource) {
	r.ID = d.ID
	r.Name = d.Name
	r.Type = d.Type
	r.URL = d.URL
	r.ExtractedContent = d.ExtractedContent
	r.RawContent = d.RawContent
	r.Status = d.Status
	r.OwnerID = d.OwnerID
	r.CreatedAt = d.CreatedAt
	r.UpdatedAt = d.UpdatedAt
}
