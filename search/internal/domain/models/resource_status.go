package models

import (
	"github.com/google/uuid"
)

type ResourceStatus string

const (
	ResourceStatusCompleted  ResourceStatus = "completed"
	ResourceStatusProcessing ResourceStatus = "processing"
	ResourceStatusFailed     ResourceStatus = "failed"
)

type ResourceStatusUpdate struct {
	ResourceID uuid.UUID      `json:"resource_id"`
	Status     ResourceStatus `json:"status"`
}
