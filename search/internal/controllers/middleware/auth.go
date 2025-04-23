package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/gin-gonic/gin"
)

// Constants for context keys
const (
	UserIDKey    string = "user_id"
	UserNameKey  string = "user_name"
	UserRolesKey string = "user_roles"
)

// AuthMiddlewareConfig holds necessary configuration for Keycloak authentication
type AuthMiddlewareConfig struct {
	Host         string
	Port         string
	Realm        string
	ClientID     string
	ClientSecret string
}

// AuthMiddleware provides JWT validation with Keycloak
type AuthMiddleware struct {
	keycloak    *gocloak.GoCloak
	config      *AuthMiddlewareConfig
	publicKey   string
	publicKeyMu sync.RWMutex
	lastFetched time.Time
}

// NewAuthMiddleware creates a new middleware instance
func NewAuthMiddleware(config *AuthMiddlewareConfig) *AuthMiddleware {
	keycloakURL := fmt.Sprintf("http://%s:%s", config.Host, config.Port)
	return &AuthMiddleware{
		keycloak: gocloak.NewClient(keycloakURL),
		config:   config,
	}
}

// Authenticate creates a gin handler function for Keycloak authentication
func (k *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, _, err := k.keycloak.DecodeAccessToken(ctx, tokenString, k.config.Realm)
		if err != nil {
			slog.Error("failed to decode access token", "error", err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		userID, err := token.Claims.GetSubject()
		if err != nil {
			slog.Error("failed to get subject from token", "error", err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID in token"})
			return
		}

		isValid, err := k.validateToken(ctx, tokenString)
		if err != nil || !isValid {
			slog.Error("token validation failed", "error", err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token validation failed"})
			return
		}

		userName, roles, err := k.getUserInfo(ctx, tokenString)
		if err != nil {
			slog.Error("failed to get user info", "error", err)
			// Continue anyway as we have the user ID
		}

		ctx.Set(UserIDKey, userID)
		ctx.Set(UserNameKey, userName)
		ctx.Set(UserRolesKey, roles)

		reqCtx := context.WithValue(ctx.Request.Context(), UserIDKey, userID)
		reqCtx = context.WithValue(reqCtx, UserNameKey, userName)
		reqCtx = context.WithValue(reqCtx, UserRolesKey, roles)
		ctx.Request = ctx.Request.WithContext(reqCtx)

		ctx.Next()
	}
}

// validateToken performs token introspection to ensure it's still valid
func (k *AuthMiddleware) validateToken(ctx context.Context, tokenString string) (bool, error) {
	rptResult, err := k.keycloak.RetrospectToken(ctx, tokenString, k.config.ClientID, k.config.ClientSecret, k.config.Realm)
	if err != nil {
		return false, fmt.Errorf("failed to introspect token: %w", err)
	}

	return *rptResult.Active, nil
}

// getUserInfo gets additional user information from Keycloak
func (k *AuthMiddleware) getUserInfo(ctx context.Context, tokenString string) (string, []string, error) {
	userInfo, err := k.keycloak.GetUserInfo(ctx, tokenString, k.config.Realm)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get user info: %w", err)
	}

	var roles []string

	return *userInfo.PreferredUsername, roles, nil
}

func GetUserID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(UserIDKey).(string)
	return id, ok
}

func GetUserName(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(UserNameKey).(string)
	return name, ok
}

func GetUserRoles(ctx context.Context) ([]string, bool) {
	roles, ok := ctx.Value(UserRolesKey).([]string)
	return roles, ok
}
