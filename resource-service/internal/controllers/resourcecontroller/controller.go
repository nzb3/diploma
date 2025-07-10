package resourcecontroller

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/nzb3/diploma/resource-service/internal/controllers"
	"github.com/nzb3/diploma/resource-service/internal/controllers/middleware"
	"github.com/nzb3/diploma/resource-service/internal/domain/models/resourcemodel"
)

const (
	DefaultLimit  = 10
	DefaultOffset = 0
)

type resourceService interface {
	SaveUsersResource(ctx context.Context, userID uuid.UUID, content []byte, resourceType resourcemodel.ResourceType, name, url string) (<-chan resourcemodel.Resource, <-chan resourcemodel.ResourceStatusUpdate, <-chan error)
	GetUsersResources(ctx context.Context, userID uuid.UUID, limit, offset int) ([]resourcemodel.Resource, error)
	GetUsersResourceByID(ctx context.Context, userID uuid.UUID, resourceID uuid.UUID) (resourcemodel.Resource, error)
	DeleteUsersResource(ctx context.Context, userID uuid.UUID, resourceID uuid.UUID) error
	UpdateUsersResource(ctx context.Context, userID uuid.UUID, resourceID uuid.UUID, name *string, content *[]byte) (resourcemodel.Resource, error)
}

type Controller struct {
	service resourceService
}

func NewController(service resourceService) *Controller {
	c := &Controller{
		service: service,
	}
	slog.Debug("Initialized resource controller")
	return c
}

func (c *Controller) RegisterRoutes(router *gin.RouterGroup) {
	slog.Info("Registering resource routes")
	resourceGroup := router.Group("/resources", middleware.RequestLogger())
	{
		resourceGroup.POST("/", middleware.SSEHeadersMiddleware(), c.SaveResource())
		resourceGroup.PATCH("/:id", c.UpdateResource())
		resourceGroup.GET("/", c.GetResources())
		resourceGroup.GET("/:id", c.GetResourceByID())
		resourceGroup.DELETE("/:id", c.DeleteResource())
	}
}

// SaveResource godoc
// @Summary      Create a new resource
// @Description  Creates a new resource for the authenticated user. Returns the created resource and status updates via SSE.
// @Tags         resources
// @Accept       json
// @Produce      json
// @Param        request  body      SaveResourceRequest  true  "Resource creation payload"
// @Success      200      {object}  SSEResourceEvent    "Resource created event (SSE)"
// @Failure      400      {object}  ErrorResponse       "Invalid user id or request body"
// @Failure      500      {object}  ErrorResponse       "Internal server error"
// @Security     ApiKeyAuth
// @Router       /resources [post]
func (c *Controller) SaveResource() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		slog.Info("Handling save resource request",
			"client", ctx.ClientIP(),
			"content_type", ctx.ContentType())

		req, ok := controllers.ValidateRequest[SaveResourceRequest](ctx)
		if !ok {
			slog.Warn("Invalid save request")
			return
		}

		userID, ok := controllers.GetUserID(ctx)
		if !ok {
			slog.Warn("Invalid user id")
			c.respondWithError(ctx, http.StatusBadRequest, "Invalid user id")
			return
		}

		resourceCh, statusUpdateCh, errCh := c.service.SaveUsersResource(ctx, userID, req.Content, resourcemodel.ResourceType(req.Type), req.Name, req.URL)

		ctx.Stream(func(w io.Writer) bool {
			select {
			case resource, ok := <-resourceCh:
				return c.handleResourceEvent(ctx, resource, ok)
			case statusUpdate, ok := <-statusUpdateCh:
				return c.handleStatusUpdateEvent(ctx, statusUpdate, ok)
			case err := <-errCh:
				return c.handleErrorEvent(ctx, err, ok)
			case <-ctx.Done():
				slog.Warn("Client disconnected", "client", ctx.ClientIP())
				return false
			}
		})
	}
}

// UpdateResource godoc
// @Summary      Update a resource
// @Description  Updates the name or content of a resource for the authenticated user.
// @Tags         resources
// @Accept       json
// @Produce      json
// @Param        id       path      string                true   "Resource ID (UUID)"
// @Param        request  body      UpdateResourceRequest true   "Fields to update"
// @Success      200      {object}  UpdateResourceResponse
// @Failure      400      {object}  ErrorResponse         "Invalid user id, resource id, or request body"
// @Failure      404      {object}  ErrorResponse         "Resource not found"
// @Failure      500      {object}  ErrorResponse         "Internal server error"
// @Security     ApiKeyAuth
// @Router       /resources/{id} [patch]
func (c *Controller) UpdateResource() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var pathReq GetResourceByIDRequest
		if err := ctx.ShouldBindUri(&pathReq); err != nil {
			slog.Error("Error parsing resource ID", "err", err)
			c.respondWithError(ctx, http.StatusBadRequest, "invalid resource ID")
			return
		}

		userID, ok := controllers.GetUserID(ctx)
		if !ok {
			slog.Warn("Invalid user id")
			c.respondWithError(ctx, http.StatusBadRequest, "Invalid user id")
			return
		}

		var req UpdateResourceRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			slog.Error("Error parsing request", "err", err)
			c.respondWithError(ctx, http.StatusBadRequest, err.Error())
			return
		}

		resource, err := c.service.UpdateUsersResource(ctx, userID, pathReq.ID, req.Name, req.Content)
		if err != nil {
			slog.Warn("Failed to update resource", "error", err)
			c.respondWithError(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		response := UpdateResourceResponse{Resource: resource}
		ctx.JSON(http.StatusOK, response)
	}
}

// GetResources godoc
// @Summary      Get list of user resources
// @Description  Returns a paginated list of resources belonging to the authenticated user.
// @Tags         resources
// @Accept       json
// @Produce      json
// @Param        limit   query     int     false  "Maximum number of resources to return"  minimum(1)  default(10)
// @Param        offset  query     int     false  "Number of resources to skip before starting to collect the result set"  minimum(0)  default(0)
// @Success      200     {object}  GetResourcesResponse
// @Failure      400     {object}  ErrorResponse  "Invalid user id or bad request"
// @Failure      500     {object}  ErrorResponse  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /resources [get]
func (c *Controller) GetResources() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		slog.Info("Fetching resources list")

		userID, ok := controllers.GetUserID(ctx)
		if !ok {
			slog.Warn("Invalid user id")
			c.respondWithError(ctx, http.StatusBadRequest, "Invalid user id")
			return
		}

		limit, offset := getPaginationParams(ctx)

		resources, err := c.service.GetUsersResources(ctx, userID, limit, offset)
		if err != nil {
			slog.Error("Failed to retrieve resources", "error", err)
			c.respondWithError(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		response := GetResourcesResponse{
			Resources: resources,
			Count:     len(resources),
		}

		slog.Info("Successfully fetched resources", "count", len(resources))
		ctx.JSON(http.StatusOK, response)
	}
}

// GetResourceByID godoc
// @Summary      Get a resource by ID
// @Description  Returns a single resource by its ID for the authenticated user.
// @Tags         resources
// @Accept       json
// @Produce      json
// @Param        id      path      string  true   "Resource ID (UUID)"
// @Success      200     {object}  GetResourceByIDResponse
// @Failure      400     {object}  ErrorResponse  "Invalid user id or resource id"
// @Failure      404     {object}  ErrorResponse  "Resource not found"
// @Failure      500     {object}  ErrorResponse  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /resources/{id} [get]
func (c *Controller) GetResourceByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, ok := controllers.GetUserID(ctx)
		if !ok {
			slog.Warn("Invalid user id")
			c.respondWithError(ctx, http.StatusBadRequest, "Invalid user id")
			return
		}

		var req GetResourceByIDRequest
		if err := ctx.ShouldBindUri(&req); err != nil {
			slog.Error("Invalid resource ID format", "error", err)
			c.respondWithError(ctx, http.StatusBadRequest, "invalid resource ID")
			return
		}

		slog.Info("Processing get resource request",
			"resource_id", req.ID,
			"client", ctx.ClientIP())

		resource, err := c.service.GetUsersResourceByID(ctx, userID, req.ID)
		if err != nil {
			slog.Error("Failed to retrieve resource",
				"resource_id", req.ID,
				"error", err)
			c.respondWithError(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		response := GetResourceByIDResponse{Resource: resource}
		slog.Info("Successfully fetched resource")
		ctx.JSON(http.StatusOK, response)
	}
}

// DeleteResource godoc
// @Summary      Delete a resource
// @Description  Deletes a resource by its ID for the authenticated user.
// @Tags         resources
// @Accept       json
// @Produce      json
// @Param        id    path      string  true   "Resource ID (UUID)"
// @Success      200   {object}  DeleteResourceResponse
// @Failure      400   {object}  ErrorResponse  "Invalid user id or resource id"
// @Failure      404   {object}  ErrorResponse  "Resource not found"
// @Failure      500   {object}  ErrorResponse  "Internal server error"
// @Security     ApiKeyAuth
// @Router       /resources/{id} [delete]
func (c *Controller) DeleteResource() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID, ok := controllers.GetUserID(ctx)
		if !ok {
			slog.Warn("Invalid user id")
			c.respondWithError(ctx, http.StatusBadRequest, "Invalid user id")
			return
		}

		var req DeleteResourceRequest
		if err := ctx.ShouldBindUri(&req); err != nil {
			slog.Error("Invalid resource ID format", "error", err)
			c.respondWithError(ctx, http.StatusBadRequest, "invalid resource ID")
			return
		}

		slog.Info("Processing delete request",
			"resource_id", req.ID,
			"client", ctx.ClientIP())

		if err := c.service.DeleteUsersResource(ctx, userID, req.ID); err != nil {
			slog.Error("Failed to delete resource",
				"resource_id", req.ID,
				"error", err)
			c.respondWithError(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		response := DeleteResourceResponse{Message: "Resource deleted successfully"}
		slog.Info("Resource deleted successfully", "resource_id", req.ID)
		ctx.JSON(http.StatusOK, response)
	}
}

// SSE Event Handlers
func (c *Controller) handleResourceEvent(ctx *gin.Context, resource resourcemodel.Resource, ok bool) bool {
	if !ok {
		return false
	}

	slog.Info("Sending resource", "resource_id", resource.ID)
	event := SSEResourceEvent{Resource: resource}
	controllers.SendSSEEvent(ctx, "resource", event)
	return false
}

func (c *Controller) handleStatusUpdateEvent(ctx *gin.Context, update resourcemodel.ResourceStatusUpdate, ok bool) bool {
	if !ok {
		slog.Debug("Resource channel closed")
		return false
	}

	slog.Info("Sending status update", "resource_id", update.ResourceID, "status", update.Status)

	event := SSEStatusUpdateEvent{
		ResourceID: update.ResourceID,
		Status:     update.Status,
	}
	controllers.SendSSEEvent(ctx, "status_update", event)

	if update.Status == resourcemodel.ResourceStatusCompleted {
		c.sendCompletionEvent(ctx, update.ResourceID)
		return false
	}
	return true
}

func (c *Controller) handleErrorEvent(ctx *gin.Context, err error, ok bool) bool {
	if ok {
		slog.Error("Resource processing error", "error", err)
		event := SSEErrorEvent{Error: err.Error()}
		controllers.SendSSEEvent(ctx, "error", event)
	}
	return false
}

func (c *Controller) sendCompletionEvent(ctx *gin.Context, id uuid.UUID) {
	slog.Info("Resource processing completed", "resource_id", id)
	event := SSECompletionEvent{ResourceID: id}
	controllers.SendSSEEvent(ctx, "completed", event)
}

func (c *Controller) respondWithError(ctx *gin.Context, statusCode int, message string) {
	response := ErrorResponse{Error: message}
	ctx.JSON(statusCode, response)
}

func getPaginationParams(ctx *gin.Context) (limit, offset int) {
	limitStr := ctx.Query("limit")

	if limitStr == "" {
		limit = DefaultLimit
	} else {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l <= 0 {
			limit = DefaultLimit
		} else {
			limit = l
		}
	}

	offsetStr := ctx.Query("offset")

	if offsetStr == "" {
		offset = DefaultOffset
	} else {
		o, err := strconv.Atoi(offsetStr)
		if err != nil || o < 0 {
			offset = DefaultOffset
		} else {
			offset = o
		}
	}

	return limit, offset
}
