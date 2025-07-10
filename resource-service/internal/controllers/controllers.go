package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

const (
	UserIDKey    string = "user_id"
	UserNameKey  string = "user_name"
	UserRolesKey string = "user_roles"
)

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return uuid.Nil, false
	}

	uuidID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, false
	}

	return uuidID, ok
}

func GetUserName(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(UserNameKey).(string)
	return name, ok
}

func GetUserRoles(ctx context.Context) ([]string, bool) {
	roles, ok := ctx.Value(UserRolesKey).([]string)
	return roles, ok
}
