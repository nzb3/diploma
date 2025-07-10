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

type Controller interface {
	RegisterRoutes(router *gin.RouterGroup)
}
