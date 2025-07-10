package models

import (
	"errors"

	"github.com/google/uuid"
)

type Embedding struct {
	UUID         uuid.UUID              `gorm:"type:uuid;primaryKey" json:"uuid"`
	CollectionID uuid.UUID              `gorm:"type:uuid" json:"collection_id"`
	Embedding    []float32              `gorm:"type:vector(1024)" json:"embedding"`
	Document     string                 `gorm:"type:text" json:"document"`
	Cmetadata    map[string]interface{} `gorm:"type:jsonb" json:"cmetadata"`
}

type EmbeddingValidationError error

var (
	ErrorMissingEmbeddingUUID     EmbeddingValidationError = errors.New("embedding uuid is missing")
	ErrorMissingCollectionID      EmbeddingValidationError = errors.New("collection id is missing")
	ErrorMissingEmbeddingDocument EmbeddingValidationError = errors.New("document is missing")
)

func (e *Embedding) Validate() error {
	if e.UUID == uuid.Nil {
		return ErrorMissingEmbeddingUUID
	}
	if e.CollectionID == uuid.Nil {
		return ErrorMissingCollectionID
	}
	if e.Document == "" {
		return ErrorMissingEmbeddingDocument
	}
	return nil
}

func (e *Embedding) TableName() string {
	return "embeddings"
}
