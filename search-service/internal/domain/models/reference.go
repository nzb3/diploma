package models

import (
	"github.com/google/uuid"
)

type Reference struct {
	ResourceID uuid.UUID `json:"resource_id"`
	Content    string    `json:"content"`
	Score      float32   `json:"score"`
}
