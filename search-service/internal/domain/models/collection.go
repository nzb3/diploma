package models

import (
	"errors"

	"github.com/google/uuid"
)

type Collection struct {
	UUID      uuid.UUID              `gorm:"type:uuid;primaryKey" json:"uuid"`
	Name      string                 `gorm:"type:varchar(255)" json:"name"`
	Cmetadata map[string]interface{} `gorm:"type:jsonb" json:"cmetadata"`
}

type CollectionValidationError error

var (
	ErrorMissingCollectionName CollectionValidationError = errors.New("collection name is missing")
	ErrorMissingCollectionUUID CollectionValidationError = errors.New("collection uuid is missing")
)

func (c *Collection) Validate() error {
	if c.UUID == uuid.Nil {
		return ErrorMissingCollectionUUID
	}
	if c.Name == "" {
		return ErrorMissingCollectionName
	}
	return nil
}

func (c *Collection) TableName() string {
	return "collections"
}
