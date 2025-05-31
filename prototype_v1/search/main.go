package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	e, err := NewEmbedder()
	if err != nil {
		slog.Error("failed to create e", "error", err)
		os.Exit(1)
	}

	llm, err := NewGenerator()
	if err != nil {
		slog.Error("failed to create LLM", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	store, err := NewStorage(ctx, e, llm)
	if err != nil {
		slog.Error("failed to create storage", "error", err)
		os.Exit(1)
	}

	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept", "User-Agent", "Cache-Control", "Pragma"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	router.Use(cors.New(config))

	h := NewHandler(router, store)

	router.POST("/ask", h.Ask())
	router.POST("/search", h.SemanticSearch())
	router.POST("/documents", h.SaveDocument())

	slog.Info("starting server on :8080")
	if err := router.Run("0.0.0.0:8080"); err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}

type handler struct {
	engine *gin.Engine
	store  *storage
}

func NewHandler(engine *gin.Engine, store *storage) *handler {
	return &handler{
		engine: engine,
		store:  store,
	}
}

type AskRequest struct {
	Question string `json:"question" binding:"required"`
}

type AskResponse struct {
	Answer string `json:"answer"`
}

func (h *handler) Ask() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req AskRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		answer, err := h.store.GetAnswer(ctx, req.Question)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, AskResponse{Answer: answer})
	}
}

type SearchRequest struct {
	Query      string `json:"query" binding:"required"`
	MaxResults int    `json:"max_results"`
}

type SearchResponse struct {
	Results []DocumentResult `json:"results"`
}

type DocumentResult struct {
	Content     string                 `json:"content"`
	Metadata    map[string]interface{} `json:"metadata"`
	PageContent string                 `json:"page_content"`
}

func (h *handler) SemanticSearch() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req SearchRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Set default max results if not provided
		maxResults := req.MaxResults
		if maxResults <= 0 {
			maxResults = numOfResults
		}

		docs, err := h.store.SemanticSearch(ctx, req.Query, maxResults)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		results := make([]DocumentResult, len(docs))
		for i, doc := range docs {
			results[i] = DocumentResult{
				Content:     doc.PageContent,
				PageContent: doc.PageContent,
				Metadata:    doc.Metadata,
			}
		}

		ctx.JSON(http.StatusOK, SearchResponse{Results: results})
	}
}

type SaveDocumentRequest struct {
	Content []byte `json:"content" binding:"required"`
	Type    string `json:"type" binding:"required"`
}

type SaveDocumentResponse struct {
	Success bool `json:"success"`
}

func (h *handler) SaveDocument() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req SaveDocumentRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var err error

		switch req.Type {
		case "url":
			err = h.store.PutSite(ctx, string(req.Content))
		case "text":
			err = h.store.PutText(ctx, string(req.Content))
		case "pdf":
			err = h.store.PutPDFFile(ctx, req.Content)
		default:
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid type"})
			return
		}

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, SaveDocumentResponse{Success: true})
	}
}
