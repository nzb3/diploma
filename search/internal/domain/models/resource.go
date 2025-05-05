package models

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nzb3/diploma/search/internal/validator"
)

type ResourceType string

type ResourceEvent struct {
	ID     uuid.UUID      `json:"id"`
	Status ResourceStatus `json:"status"`
}

type Resource struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Name             string         `gorm:"type:varchar(255)" json:"name"`
	Type             ResourceType   `gorm:"type:varchar(100)" json:"type"`
	URL              string         `gorm:"type:varchar(255)" json:"url,omitempty"`
	ExtractedContent string         `gorm:"type:text" json:"extracted_content"`
	RawContent       []byte         `gorm:"type:bytea" json:"raw_content"`
	ChunkIDs         []string       `gorm:"-" json:"chunk_ids,omitempty"`
	Status           ResourceStatus `gorm:"type:varchar(50)" json:"status,omitempty"`
	OwnerID          string         `gorm:"type:varchar(100)" json:"owner_id,omitempty"`
	CreatedAt        time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
}

func (r *Resource) SetStatusFailed() {
	r.Status = ResourceStatusFailed
}

func (r *Resource) SetStatusProcessing() {
	r.Status = ResourceStatusProcessing
}

func (r *Resource) SetStatusCompleted() {
	r.Status = ResourceStatusCompleted
}

func (r *Resource) Validate(validators ...validator.ValidateFunc[Resource]) error {
	var err error
	for _, fn := range validators {
		if validationErr := fn(r); validationErr != nil {
			err = errors.Join(err, validationErr)
		}
	}

	if err != nil {
		return err
	}

	return nil
}

func (r *Resource) HaveID() validator.ValidateFunc[Resource] {
	return func(r *Resource) error {
		if r.ID == uuid.Nil {
			return ValidationErrorMissingID
		}

		return nil
	}

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
		if len(r.ChunkIDs) != 0 {
			var existingEmbeddingIDs []uuid.UUID
			if err := tx.Model(&ResourceEmbedding{}).
				Where("resource_id = ?", r.ID).
				Pluck("embedding_id", &existingEmbeddingIDs).
				Error; err != nil {
				return fmt.Errorf("%s: finding existing embedding IDs: %w", op, err)
			}

			if len(existingEmbeddingIDs) > 0 {
				if err := tx.Where("uuid IN ?", existingEmbeddingIDs).Delete(&Embedding{}).Error; err != nil {
					return fmt.Errorf("%s: deleting existing embeddings: %w", op, err)
				}
			}

			if err := tx.Where("resource_id = ?", r.ID).Delete(&ResourceEmbedding{}).Error; err != nil {
				return fmt.Errorf("%s: deleting existing resource embedding associations: %w", op, err)
			}

			associations := make([]ResourceEmbedding, 0, len(r.ChunkIDs))
			for _, chunkID := range r.ChunkIDs {
				associations = append(associations, ResourceEmbedding{
					ResourceID:  r.ID,
					EmbeddingID: uuid.MustParse(chunkID),
				})
			}

			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&associations).Error; err != nil {
				return fmt.Errorf("%s: creating new resource embeddings: %w", op, err)
			}
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
