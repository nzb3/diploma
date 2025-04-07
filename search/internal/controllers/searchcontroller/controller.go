package searchcontroller

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/nzb3/diploma/search/internal/controllers"
	"github.com/nzb3/diploma/search/internal/controllers/middleware"
	"github.com/nzb3/diploma/search/internal/domain/models"
)

type searchService interface {
	GetAnswer(ctx context.Context, question string, refsChan chan<- []models.Reference) (models.SearchResult, error)
	GetAnswerStream(ctx context.Context, question string) (<-chan models.SearchResult, <-chan []models.Reference, <-chan []byte, <-chan error)
	SemanticSearch(ctx context.Context, query string) ([]models.Reference, error)
}

type Controller struct {
	searchService  searchService
	activeRequests sync.Map
}

func NewController(ss searchService) *Controller {
	return &Controller{
		searchService: ss,
	}
}

func (c *Controller) RegisterRoutes(router *gin.RouterGroup) {
	slog.Debug("Registering routes")
	askGroup := router.Group("/ask", middleware.RequestLogger())
	{
		askGroup.POST("/", middleware.SSEHeadersMiddleware(), c.createProcessMiddleware(), c.Ask())
		streamGroup := askGroup.Group("/stream")
		{
			streamGroup.GET("/", middleware.SSEHeadersMiddleware(), c.createProcessMiddleware(), c.AskStream())
			streamGroup.DELETE("/cancel/:process_id", c.CancelProcess())
		}
	}

	searchGroup := router.Group("/search")
	{
		searchGroup.POST("/", c.SemanticSearch())
	}
}

type AskRequest struct {
	Question string `json:"question" binding:"required"`
}

type AskResponse struct {
	Result models.SearchResult `json:"result"`
}

func (c *Controller) Ask() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		slog.Info("Handling Ask request")
		var req AskRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			slog.Error("Error binding request", "error", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		processID, err := getProcessIDFromContext(ctx)
		if err != nil {
			slog.Error("Error getting process ID check createProcessMiddleware", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start process"})
			return
		}

		refsChan := make(chan []models.Reference, 1)
		go func() {
			defer close(refsChan)

			ctx.Stream(func(w io.Writer) bool {
				select {
				case <-ctx.Done():
					return false
				case refs := <-refsChan:
					return c.handleReferences(ctx, processID, refs)
				}
			})
		}()
		slog.Debug("Processing question", "question", req.Question)
		answer, err := c.searchService.GetAnswer(ctx, req.Question, refsChan)

		if err != nil {
			slog.Error("Error getting answer", "error", err, "question", req.Question)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		slog.Info("Successfully processed request", "question", req.Question)
		ctx.JSON(http.StatusOK, AskResponse{Result: answer})
	}
}

func (c *Controller) AskStream() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		slog.Info("Initializing stream request")
		question := ctx.Query("question")
		if question == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "question is required"})
			return
		}
		slog.Info("Processing question", "question", question)

		processID, err := getProcessIDFromContext(ctx)
		if err != nil {
			slog.Error("Error getting process ID check createProcessMiddleware", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start process"})
			return
		}

		slog.Info("Starting stream processing",
			"process_id", processID,
			"question", question,
			"client", ctx.ClientIP())

		resultChan, referencesChan, chunkChan, errChan := c.searchService.GetAnswerStream(ctx, question)

		ctx.Stream(func(w io.Writer) bool {
			select {
			case chunk := <-chunkChan:
				return c.handleChunk(ctx, processID, chunk)
			case references := <-referencesChan:
				return c.handleReferences(ctx, processID, references)
			case result := <-resultChan:
				return c.handleResult(ctx, processID, result)
			case err := <-errChan:
				return c.handleError(ctx, processID, err)
			case <-ctx.Done():
				return c.handleCancellationEvent(ctx, processID, ctx.Err())
			}
		})
	}
}

func getProcessIDFromContext(ctx *gin.Context) (uuid.UUID, error) {
	value, ok := ctx.Get("process_id")
	if !ok {
		return uuid.Nil, errors.New("process_id not found in context")
	}

	processID, ok := value.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("process_id not found in context")
	}

	if processID == uuid.Nil {
		return uuid.Nil, errors.New("process_id not found in context")
	}

	return processID, nil
}

func (c *Controller) createProcessMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		processID := uuid.New()
		cancelCtx, cancel := context.WithCancel(ctx.Request.Context())
		c.activeRequests.Store(processID, cancel)

		ctx.Request = ctx.Request.WithContext(cancelCtx)

		ctx.Set("process_id", processID)

		slog.Debug("Created new process context",
			"process_id", processID,
			"active_requests", c.activeRequestsCount())

		ctx.Next()
	}
}

func (c *Controller) cleanupProcess(processID uuid.UUID) {
	if cancel, loaded := c.activeRequests.LoadAndDelete(processID); loaded {
		slog.Debug("Cleaning up process", "process_id", processID)
		cancel.(context.CancelFunc)()
	}
}

func (c *Controller) handleReferences(ctx *gin.Context, processID uuid.UUID, references []models.Reference) bool {
	slog.Debug("Processing reference",
		"process_id", processID,
		"references", references)
	controllers.SendSSEEvent(ctx, "resources", gin.H{
		"process_id": processID,
		"references": references,
		"complete":   false,
	})
	return true
}

func (c *Controller) handleChunk(ctx *gin.Context, processID uuid.UUID, chunk []byte) bool {
	slog.Debug("Processing chunk", "process_id", processID, "chunk_size", len(chunk))
	controllers.SendSSEEvent(ctx, "chunk", gin.H{
		"process_id": processID.String(),
		"content":    string(chunk),
		"complete":   false,
	})
	return true
}

func (c *Controller) handleResult(ctx *gin.Context, processID uuid.UUID, result models.SearchResult) bool {
	slog.Info("Finalizing stream processing", "process_id", processID, "result", result)

	controllers.SendSSEEvent(ctx, "complete", gin.H{
		"process_id": processID.String(),
		"result":     result,
		"complete":   true,
	})

	slog.Debug("Sent final result", "process_id", processID, "result", result)
	return false
}

func (c *Controller) handleError(ctx *gin.Context, processID uuid.UUID, err error) bool {
	slog.Error("Stream processing error", "process_id", processID, "error", err)

	controllers.SendSSEEvent(ctx, "error", gin.H{
		"process_id": processID.String(),
		"error":      err.Error(),
	})
	slog.Error("Stream error occurred", "process_id", processID, "error", err)
	c.cleanupProcess(processID)
	return false
}

func (c *Controller) handleCancellationEvent(ctx *gin.Context, processID uuid.UUID, err error) bool {
	slog.Warn("Stream processing cancelled", "process_id", processID, "reason", err)

	controllers.SendSSEEvent(ctx, "cancelled", gin.H{
		"process_id": processID.String(),
		"message":    "Request cancelled by user",
	})

	slog.Info("Cancellation completed", "process_id", processID, "client", ctx.ClientIP())

	return false
}

func (c *Controller) CancelProcess() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		processID := ctx.Param("process_id")
		slog.Info("Processing cancellation request",
			"process_id", processID,
			"client", ctx.ClientIP())

		uuidID, err := uuid.Parse(processID)
		if err != nil {
			slog.Warn("Invalid process ID format",
				"input", processID,
				"error", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid process id"})
			return
		}

		if cancel, ok := c.activeRequests.Load(uuidID); ok {
			slog.Debug("Found active process to cancel", "process_id", uuidID)
			cancel.(context.CancelFunc)()
			ctx.JSON(http.StatusOK, gin.H{"message": "Cancellation requested"})
		} else {
			slog.Warn("Process not found for cancellation", "process_id", uuidID)
			ctx.JSON(http.StatusNotFound, gin.H{"error": "process not found"})
		}
	}
}

type SearchRequest struct {
	Query      string `json:"query" binding:"required"`
	MaxResults int    `json:"max_results"`
}

type SearchResponse struct {
	References []models.Reference `json:"references"`
}

func (c *Controller) SemanticSearch() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		slog.Info("Handling semantic search request")
		var req SearchRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			slog.Error("Invalid search request", "error", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		slog.Debug("Executing semantic search",
			"query", req.Query,
			"max_results", req.MaxResults)

		references, err := c.searchService.SemanticSearch(ctx, req.Query)
		if err != nil {
			slog.Error("Semantic search failed",
				"error", err,
				"query", req.Query)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		slog.Info("Semantic search completed",
			"query", req.Query,
			"results_count", len(references))
		ctx.JSON(http.StatusOK, SearchResponse{References: references})
	}
}

func (c *Controller) activeRequestsCount() int {
	count := 0
	c.activeRequests.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}
