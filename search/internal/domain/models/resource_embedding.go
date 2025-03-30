package models

import (
	"errors"

	"github.com/google/uuid"
)

type ResourceEmbedding struct {
	ResourceID  uuid.UUID `gorm:"type:uuid;primaryKey;column:resource_id" json:"resource_id"` // Matches schema spelling
	EmbeddingID uuid.UUID `gorm:"type:uuid;primaryKey;unique;column:embedding_id" json:"embedding_id"`
}

type ResourceEmbeddingValidationError error

var (
	ErrorMissingResourceID  ResourceEmbeddingValidationError = errors.New("resource id is missing")
	ErrorMissingEmbeddingID ResourceEmbeddingValidationError = errors.New("embedding id is missing")
)

func (re *ResourceEmbedding) Validate() error {
	if re.ResourceID == uuid.Nil {
		return ErrorMissingResourceID
	}
	if re.EmbeddingID == uuid.Nil {
		return ErrorMissingEmbeddingID
	}
	return nil
}

func (re *ResourceEmbedding) TableName() string {
	return "resource_embedding"
}
