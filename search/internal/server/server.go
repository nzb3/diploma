package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Controller interface {
	RegisterRoutes(router *gin.RouterGroup)
}

func NewServer(ctx context.Context, router *gin.Engine, cfg *Config, controllers ...Controller) *http.Server {
	api := router.Group("/api")
	v1 := api.Group("/v1")

	for _, controller := range controllers {
		controller.RegisterRoutes(v1)
	}

	s := &http.Server{
		Addr:              ":" + cfg.HTTP.Port,
		Handler:           router,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		ReadHeaderTimeout: cfg.HTTP.ReadHeaderTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
		MaxHeaderBytes:    cfg.HTTP.MaxHeaderBytes,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			cfg.HTTP.ShutdownTimeout,
		)
		defer cancel()

		if err := s.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown error", zap.Error(err))
			panic(fmt.Errorf("server shutdown error: %w", err))
		}
	}()

	return s
}
