package resourcecontroller

import (
	"github.com/google/uuid"

	"github.com/nzb3/diploma/resource-service/internal/domain/models/resourcemodel"
)

// SaveResourceRequest represents the payload for creating a resource.
// swagger:model SaveResourceRequest
type SaveResourceRequest struct {
	// Resource content (binary data)
	// Required: true
	Content []byte `json:"content" binding:"required"`
	// Resource type (e.g. document, image)
	// Required: true
	Type string `json:"type" binding:"required"`
	// Optional resource name
	Name string `json:"name,omitempty"`
	// Optional resource URL
	URL string `json:"url,omitempty"`
}

// UpdateResourceRequest represents the payload for updating a resource.
// Only provided fields will be updated.
// swagger:model UpdateResourceRequest
type UpdateResourceRequest struct {
	// New resource name (optional)
	Name *string `json:"name,omitempty"`
	// New resource content (optional, binary)
	Content *[]byte `json:"content,omitempty"`
}

// GetResourceByIDRequest represents the URI parameter for getting a resource by ID.
// swagger:model GetResourceByIDRequest
type GetResourceByIDRequest struct {
	// Resource ID (UUID)
	// in: path
	// Required: true
	ID uuid.UUID `uri:"id" binding:"required"`
}

// DeleteResourceRequest represents the URI parameter for deleting a resource by ID.
// swagger:model DeleteResourceRequest
type DeleteResourceRequest struct {
	// Resource ID (UUID)
	// in: path
	// Required: true
	ID uuid.UUID `uri:"id" binding:"required"`
}

// SaveResourceResponse represents the response for resource creation.
// swagger:model SaveResourceResponse
type SaveResourceResponse struct {
	// The created resource
	Resource resourcemodel.Resource `json:"resource"`
}

// UpdateResourceResponse represents the response for resource update.
// swagger:model UpdateResourceResponse
type UpdateResourceResponse struct {
	// The updated resource
	Resource resourcemodel.Resource `json:"resource"`
}

// GetResourcesResponse represents a paginated list of resources.
// swagger:model GetResourcesResponse
type GetResourcesResponse struct {
	// List of resources
	Resources []resourcemodel.Resource `json:"resources"`
	// Total count of resources
	Count int `json:"count"`
}

// GetResourceByIDResponse represents the response for getting a resource by ID.
// swagger:model GetResourceByIDResponse
type GetResourceByIDResponse struct {
	// The resource
	Resource resourcemodel.Resource `json:"resource"`
}

// DeleteResourceResponse represents the response for resource deletion.
// swagger:model DeleteResourceResponse
type DeleteResourceResponse struct {
	// Deletion result message
	Message string `json:"message"`
}

// ErrorResponse represents a standard error response.
// swagger:model ErrorResponse
type ErrorResponse struct {
	// Error message
	Error string `json:"error"`
}

// SSEResourceEvent represents an SSE event with a resource payload.
// swagger:model SSEResourceEvent
type SSEResourceEvent struct {
	// The resource data
	Resource resourcemodel.Resource `json:"resource"`
}

// SSEStatusUpdateEvent represents an SSE event for resource status updates.
// swagger:model SSEStatusUpdateEvent
type SSEStatusUpdateEvent struct {
	// Resource ID (UUID)
	ResourceID uuid.UUID `json:"resource_id"`
	// New status
	Status resourcemodel.ResourceStatus `json:"status"`
}

// SSECompletionEvent represents an SSE event for resource completion.
// swagger:model SSECompletionEvent
type SSECompletionEvent struct {
	// Resource ID (UUID)
	ResourceID uuid.UUID `json:"resource_id"`
}

// SSEErrorEvent represents an SSE event for errors.
// swagger:model SSEErrorEvent
type SSEErrorEvent struct {
	// Error message
	Error string `json:"error"`
}
