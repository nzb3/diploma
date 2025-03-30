package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ValidateRequest[T any](ctx *gin.Context) (*T, bool) {
	var req T
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, false
	}
	return &req, true
}

func SendSSEEvent(ctx *gin.Context, event string, data interface{}) {
	ctx.SSEvent(event, data)
	ctx.Writer.Flush()
}

func SetSSEHeaders(ctx *gin.Context) {
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
}
