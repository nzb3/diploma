package resourcecontroller

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/nzb3/diploma/search/internal/controllers"

	"github.com/nzb3/diploma/search/internal/controllers/middleware"
	"github.com/nzb3/diploma/search/internal/domain/models"
)

type resourceService interface {
	SaveResource(ctx context.Context, resource models.Resource) (<-chan models.Resource, <-chan error)
	GetResources(ctx context.Context) ([]models.Resource, error)
	GetResourceByID(ctx context.Context, resourceID uuid.UUID) (models.Resource, error)
	DeleteResource(ctx context.Context, resourceID uuid.UUID) error
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
		resourceGroup.GET("/", c.GetResources())
		resourceGroup.GET("/:id", c.GetResourceByID())
		resourceGroup.DELETE("/:id", c.DeleteResource())
	}
}

type SaveDocumentRequest struct {
	Content []byte `json:"content" binding:"required"`
	Type    string `json:"type" binding:"required"`
	Name    string `json:"name"`
}

type SaveDocumentResponse struct {
	Success bool `json:"success"`
}

func (c *Controller) SaveResource() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		slog.Info("Handling save resource request",
			"client", ctx.ClientIP(),
			"content_type", ctx.ContentType())

		req, ok := controllers.ValidateRequest[SaveDocumentRequest](ctx)
		if !ok {
			slog.Warn("Invalid save request")
			return
		}

		resourceChan, errChan := c.initProcessingChannels(ctx, req)
		c.handleResourceStream(ctx, resourceChan, errChan)
	}
}

func (c *Controller) GetResources() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		slog.Info("Fetching resources list")
		resources, err := c.service.GetResources(ctx)
		if err != nil {
			slog.Error("Failed to retrieve resources",
				"error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		slog.Info("Successfully fetched resources",
			"count", len(resources))
		ctx.JSON(http.StatusOK, resources)
	}
}

func (c *Controller) DeleteResource() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resourceID := ctx.Param("id")
		slog.Info("Processing delete request",
			"resource_id", resourceID,
			"client", ctx.ClientIP())

		uuidID, err := uuid.Parse(resourceID)
		if err != nil {
			slog.Error("Invalid resource ID format",
				"input", resourceID,
				"error", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource ID"})
			return
		}

		if err := c.service.DeleteResource(ctx, uuidID); err != nil {
			slog.Error("Failed to delete resource",
				"resource_id", uuidID,
				"error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		slog.Info("Resource deleted successfully",
			"resource_id", uuidID)
		ctx.JSON(http.StatusOK, gin.H{"message": "Resource deleted successfully"})
	}
}

func (c *Controller) GetResourceByID() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		resourceID := ctx.Param("id")
		slog.Info("Processing get resource request",
			"resource_id", resourceID,
			"client", ctx.ClientIP(),
		)

		uuidID, err := uuid.Parse(resourceID)
		if err != nil {
			slog.Error("Invalid resource ID format",
				"input", resourceID,
				"error", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid resource ID"})
			return
		}

		resource, err := c.service.GetResourceByID(ctx, uuidID)
		if err != nil {
			slog.Error("Failed to retrieve resource",
				"resource_id", uuidID,
				"error", err,
			)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		slog.Info("Successfully fetched resource")
		ctx.JSON(http.StatusOK, resource)
	}
}

func (c *Controller) initProcessingChannels(
	ctx *gin.Context,
	req *SaveDocumentRequest,
) (<-chan models.Resource, <-chan error) {
	resource := models.Resource{
		RawContent: req.Content,
		Type:       models.ResourceType(req.Type),
		Name:       req.Name,
	}

	slog.Debug("Starting resource processing",
		"resource_type", req.Type,
		"content_size", len(req.Content))

	resourceChan, errChan := c.service.SaveResource(ctx, resource)
	return resourceChan, errChan
}

func (c *Controller) handleResourceStream(
	ctx *gin.Context,
	resourceChan <-chan models.Resource,
	errChan <-chan error,
) {
	ctx.Stream(func(w io.Writer) bool {
		select {
		case res, ok := <-resourceChan:
			return c.handleResourceUpdate(ctx, res, ok)

		case err, ok := <-errChan:
			return c.handleResourceError(ctx, err, ok)

		case <-ctx.Request.Context().Done():
			slog.Warn("Client disconnected", "client", ctx.ClientIP())
			return false
		}
	})
}

func (c *Controller) handleResourceUpdate(
	ctx *gin.Context,
	res models.Resource,
	ok bool,
) bool {
	if !ok {
		slog.Debug("Resource channel closed")
		return false
	}

	slog.Info("Sending status update",
		"resource_id", res.ID,
		"status", res.Status)

	controllers.SendSSEEvent(ctx, "status_update", gin.H{
		"id":     res.ID,
		"status": res.Status,
	})

	if res.Status == models.StatusProcessed {
		c.sendCompletionEvent(ctx, res.ID)
		return false
	}
	return true
}

func (c *Controller) sendCompletionEvent(ctx *gin.Context, id uuid.UUID) {
	slog.Info("Resource processing completed", "resource_id", id)
	controllers.SendSSEEvent(ctx, "completed", gin.H{
		"id": id,
	})
}

func (c *Controller) handleResourceError(
	ctx *gin.Context,
	err error,
	ok bool,
) bool {
	if ok {
		slog.Error("Resource processing error",
			"error", err,
			"resource_type", ctx.Request.Context().Value("type"))
		controllers.SendSSEEvent(ctx, "error", gin.H{
			"error": err.Error(),
		})
	}
	return false
}
