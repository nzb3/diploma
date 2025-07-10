package resourcemodel

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/nzb3/diploma/resource-service/internal/validator"
)

type ResourceType string

const (
	ResourceTypeText ResourceType = "text"
	ResourceTypePDF  ResourceType = "pdf"
	ResourceTypeURL  ResourceType = "url"
)

type ResourceEvent struct {
	ID     uuid.UUID      `json:"id"`
	Status ResourceStatus `json:"status"`
}

type Resource struct {
	ID               uuid.UUID      `json:"id"`
	Name             string         `json:"name"`
	Type             ResourceType   `json:"type"`
	URL              string         `json:"url,omitempty"`
	ExtractedContent string         `json:"extracted_content"`
	RawContent       []byte         `json:"raw_content"`
	Status           ResourceStatus `json:"status,omitempty"`
	OwnerID          uuid.UUID      `json:"owner_id,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

func NewResource(opts ...ResourceOption) Resource {
	resource := &Resource{
		ID: uuid.New(),
	}
	for _, opt := range opts {
		opt(resource)
	}

	if resource.Name == "" {
		resource.SetDefaultName()
	}

	return *resource
}

func (r *Resource) SetStatusPending() {
	r.Status = ResourceStatusPending
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

func (r *Resource) HaveID() error {
	if r.ID == uuid.Nil {
		return ErrorMissingID
	}

	return nil

}

func (r *Resource) HaveName() error {
	if r.Name == "" {
		return ErrorMissingName
	}
	return nil
}

func (r *Resource) HaveValidType() error {
	switch r.Type {
	case ResourceTypeText, ResourceTypePDF, ResourceTypeURL:
		return nil
	default:
		return ErrorWrongType
	}
}

func (r *Resource) SetDefaultName() {
	rawContentStr := string(r.RawContent)
	trimContent := strings.TrimSpace(rawContentStr)
	splitContent := strings.Split(trimContent, " ")
	r.Name = strings.Join(splitContent[0:6], " ")
}

type ResourceOption func(*Resource)

func WithID(id uuid.UUID) ResourceOption {
	return func(r *Resource) {
		r.ID = id
	}
}

func WithName(name string) ResourceOption {
	return func(r *Resource) {
		r.Name = name
	}
}

func WithType(resourceType ResourceType) ResourceOption {
	return func(r *Resource) {
		r.Type = resourceType
	}
}

func WithURL(url string) ResourceOption {
	return func(r *Resource) {
		r.URL = url
	}
}

func WithExtractedContent(extractedContent string) ResourceOption {
	return func(r *Resource) {
		r.ExtractedContent = extractedContent
	}
}

func WithRawContent(rawContent []byte) ResourceOption {
	return func(r *Resource) {
		r.RawContent = rawContent
	}
}

func WithStatus(status ResourceStatus) ResourceOption {
	return func(r *Resource) {
		r.Status = status
	}
}

func WithOwnerID(ownerID uuid.UUID) ResourceOption {
	return func(r *Resource) {
		r.OwnerID = ownerID
	}
}
