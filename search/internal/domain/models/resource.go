package models

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ResourceStatus string

const (
	StatusSaved     ResourceStatus = "saved"
	StatusProcessed ResourceStatus = "processed"
)

type ResourceType string

const (
	ResourceTypeURL  ResourceType = "url"
	ResourceTypePDF  ResourceType = "pdf"
	ResourceTypeText ResourceType = "text"
)

type ResourceEvent struct {
	ID     uuid.UUID      `json:"id"`
	Status ResourceStatus `json:"status"`
}

type Resource struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name             string         `gorm:"type:varchar(255)" json:"name"`
	Type             ResourceType   `gorm:"type:varchar(100)" json:"type"`
	Source           string         `gorm:"type:varchar(255)" json:"source"`
	ExtractedContent string         `gorm:"type:text" json:"extracted_content"`
	RawContent       []byte         `gorm:"type:bytea" json:"raw_content"`
	ChunkIDs         []string       `gorm:"-" json:"chunk_ids,omitempty"`
	Status           ResourceStatus `gorm:"type:varchar(50)" json:"status,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (r *Resource) SetStatusSaved() {
	r.Status = StatusSaved
}

func (r *Resource) SetStatusProcessed() {
	r.Status = StatusProcessed
}

func (r *Resource) Validate() error {
	if r.ID == uuid.Nil {
		return ValidationErrorMissingID
	}
	if r.Name == "" {
		return ValidationErrorMissingName
	}

	if r.Type == "" {
		return ValidationErrorMissingType
	}

	if r.RawContent == nil {
		return ValidationErrorMissingRawContent
	}

	return nil
}

func (r *Resource) TableName() string {
	return "resources"
}

func (r *Resource) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}

	if r.ChunkIDs == nil {
		r.ChunkIDs = make([]string, 0)
	}

	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()

	r.SetStatusSaved()

	if len(r.ChunkIDs) > 0 {
		return tx.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(r).Error; err != nil {
				return err
			}

			associations := make([]ResourceEmbedding, 0, len(r.ChunkIDs))
			for _, chunkID := range r.ChunkIDs {
				associations = append(associations, ResourceEmbedding{
					ResourceID:  r.ID,
					EmbeddingID: uuid.MustParse(chunkID),
				})
			}

			return tx.Create(&associations).Error
		})
	}
	return nil
}

func (r *Resource) BeforeUpdate(tx *gorm.DB) error {
	const op = "Resource.BeforeUpdate"

	if r.ChunkIDs == nil {
		r.ChunkIDs = make([]string, 0)
	}

	r.UpdatedAt = time.Now()
	return tx.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("resource_id = ?", r.ID).Delete(&ResourceEmbedding{}).Error; err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		if len(r.ChunkIDs) == 0 {
			return fmt.Errorf("%s: %w", op, errors.New("no chunk ids to associate"))
		}

		associations := make([]ResourceEmbedding, 0, len(r.ChunkIDs))
		for _, chunkID := range r.ChunkIDs {
			associations = append(associations, ResourceEmbedding{
				ResourceID:  r.ID,
				EmbeddingID: uuid.MustParse(chunkID),
			})
		}

		if err := tx.Create(&associations).Error; err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}

		return nil
	})
}

func (r *Resource) BeforeDelete(tx *gorm.DB) error {
	const op = "Resource.BeforeDelete"

	var embeddingIDs []uuid.UUID
	if err := tx.Model(&ResourceEmbedding{}).
		Where("resource_id = ?", r.ID).
		Pluck("embedding_id", &embeddingIDs).
		Error; err != nil {
		return err
	}

	slog.Info("embedding ids", "op", op, "ids", embeddingIDs)

	if len(embeddingIDs) > 0 {
		if err := tx.Where("uuid IN ?", embeddingIDs).Delete(&Embedding{}).Error; err != nil {
			return err
		}

		if err := tx.Where("resource_id = ?", r.ID).Delete(&ResourceEmbedding{}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *Resource) SetDefaultName() {
	rawContentStr := string(r.RawContent)
	trimContent := strings.TrimSpace(rawContentStr)
	splitContent := strings.Split(trimContent, " ")
	r.Name = strings.Join(splitContent[0:6], " ")
}
