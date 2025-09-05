package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/nzb3/diploma/resource-service/internal/validator"

	"github.com/Nerzal/gocloak/v13"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"github.com/nzb3/diploma/resource-service/internal/configurator"
	"github.com/nzb3/diploma/resource-service/internal/controllers"
)

type AuthMiddlewareConfig struct {
	Host         string
	Port         string
	Realm        string
	ClientID     string
	ClientSecret string
}

func NewAuthMiddlewareConfig() (*AuthMiddlewareConfig, error) {
	config := new(AuthMiddlewareConfig)
	config.Host = configurator.GetString("auth.host")
	config.Port = configurator.GetString("auth.port")
	config.Realm = configurator.GetString("auth.realm")
	config.ClientID = configurator.GetString("auth.client_id")
	config.ClientSecret = configurator.GetString("auth.client_secret")

	err := validator.Validate(config)
	if err != nil {
		return nil, fmt.Errorf("invalid auth middleware config: %w", err)
	}

	return config, nil
}

type AuthMiddleware struct {
	keycloak    *gocloak.GoCloak
	config      *AuthMiddlewareConfig
	publicKey   string
	publicKeyMu sync.RWMutex
	lastFetched time.Time
}

func NewAuthMiddleware(config *AuthMiddlewareConfig) *AuthMiddleware {
	keycloakURL := fmt.Sprintf("http://%s:%s", config.Host, config.Port)
	return &AuthMiddleware{
		keycloak: gocloak.NewClient(keycloakURL),
		config:   config,
	}
}

func (k *AuthMiddleware) getToken(ctx *gin.Context) (*jwt.Token, *jwt.MapClaims, error) {
	token, claims, headersErr := k.getFromHeaders(ctx)
	if headersErr == nil {
		return token, claims, nil
	}

	token, claims, paramsErr := k.getFromParams(ctx)
	if paramsErr == nil {
		return token, claims, nil
	}

	err := errors.Join(headersErr, paramsErr)
	return nil, nil, fmt.Errorf("token was not found neither in headers nor in params: %w", err)
}

func (k *AuthMiddleware) getFromParams(ctx *gin.Context) (*jwt.Token, *jwt.MapClaims, error) {
	tokenString := ctx.Query("auth_token")
	if tokenString == "" {
		return nil, nil, errors.New("token is required")
	}

	token, claims, err := k.keycloak.DecodeAccessToken(ctx, tokenString, k.config.Realm)
	if err != nil {
		return nil, nil, err
	}
	return token, claims, nil
}

func (k *AuthMiddleware) getFromHeaders(ctx *gin.Context) (*jwt.Token, *jwt.MapClaims, error) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return nil, nil, errors.New("authorization header is required")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, nil, errors.New("invalid authorization format")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		return nil, nil, errors.New("token not found")
	}

	token, claims, err := k.keycloak.DecodeAccessToken(ctx, tokenString, k.config.Realm)
	if err != nil {
		return nil, nil, err
	}
	return token, claims, nil
}

func (k *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, _, err := k.getToken(ctx)
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

		isValid, err := k.validateToken(ctx, token.Raw)
		if err != nil || !isValid {
			slog.Error("token validation failed", "error", err)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token validation failed"})
			return
		}

		userName, roles, err := k.getUserInfo(ctx, token.Raw)
		if err != nil {
			slog.Error("failed to get user info", "error", err)
			// Continue anyway as we have the user ID
		}

		ctx.Set(controllers.UserIDKey, userID)
		ctx.Set(controllers.UserNameKey, userName)
		ctx.Set(controllers.UserRolesKey, roles)

		reqCtx := context.WithValue(ctx.Request.Context(), controllers.UserIDKey, userID)
		reqCtx = context.WithValue(reqCtx, controllers.UserNameKey, userName)
		reqCtx = context.WithValue(reqCtx, controllers.UserRolesKey, roles)
		ctx.Request = ctx.Request.WithContext(reqCtx)

		ctx.Next()
	}
}

func (k *AuthMiddleware) validateToken(ctx context.Context, tokenString string) (bool, error) {
	rptResult, err := k.keycloak.RetrospectToken(ctx, tokenString, k.config.ClientID, k.config.ClientSecret, k.config.Realm)
	if err != nil {
		return false, fmt.Errorf("failed to introspect token: %w", err)
	}

	return *rptResult.Active, nil
}

func (k *AuthMiddleware) getUserInfo(ctx context.Context, tokenString string) (string, []string, error) {
	userInfo, err := k.keycloak.GetUserInfo(ctx, tokenString, k.config.Realm)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get user info: %w", err)
	}

	var roles []string

	return *userInfo.PreferredUsername, roles, nil
}
