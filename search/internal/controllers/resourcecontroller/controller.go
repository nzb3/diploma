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
	SaveResource(ctx context.Context, resource models.Resource) (<-chan models.Resource, <-chan models.ResourceStatusUpdate, <-chan error)
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
	Content []byte `binding:"required" json:"content"`
	Type    string `binding:"required" json:"type"`
	Name    string `json:"name"`
	URL     string `json:"url"`
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

		resource := models.Resource{
			RawContent: req.Content,
			Type:       models.ResourceType(req.Type),
			Name:       req.Name,
			URL:        req.URL,
		}

		resourceCh, statusUpdateCh, errCh := c.service.SaveResource(ctx, resource)

		ctx.Stream(func(w io.Writer) bool {
			select {
			case resource, ok := <-resourceCh:
				return c.handleResource(ctx, resource, ok)
			case statusUpdate, ok := <-statusUpdateCh:
				return c.handleStatusUpdate(ctx, statusUpdate, ok)

			case err := <-errCh:
				return c.handleResourceError(ctx, err, ok)

			case <-ctx.Done():
				slog.Warn("Client disconnected", "client", ctx.ClientIP())
				return false
			}
		})
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

func (c *Controller) handleResource(ctx *gin.Context, resource models.Resource, ok bool) bool {
	if !ok {
		return false
	}

	slog.Info("Sending resource", "resource_id", resource.ID)

	controllers.SendSSEEvent(ctx, "resource", gin.H{
		"resource": resource,
	})

	return false
}

func (c *Controller) handleStatusUpdate(
	ctx *gin.Context,
	update models.ResourceStatusUpdate,
	ok bool,
) bool {
	if !ok {
		slog.Debug("Resource channel closed")
		return false
	}

	slog.Info("Sending status update", "resource_id", update.ResourceID, "status", update.Status)

	controllers.SendSSEEvent(ctx, "status_update", gin.H{
		"resource_id": update.ResourceID,
		"status":      update.Status,
	})

	if update.Status == models.ResourceStatusCompleted {
		c.sendCompletionEvent(ctx, update.ResourceID)
		return false
	}
	return true
}

func (c *Controller) sendCompletionEvent(ctx *gin.Context, id uuid.UUID) {
	slog.Info("Resource processing completed", "resource_id", id)
	controllers.SendSSEEvent(ctx, "completed", gin.H{
		"resource_id": id,
	})
}

func (c *Controller) handleResourceError(
	ctx *gin.Context,
	err error,
	ok bool,
) bool {
	if ok {
		slog.Error("Resource processing error", "error", err, "resource_type", ctx.Request.Context().Value("type"))

		controllers.SendSSEEvent(ctx, "error", gin.H{
			"error": err.Error(),
		})
	}
	return false
}
