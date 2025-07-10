package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewServer(ctx context.Context, router *gin.Engine, cfg *Config) *http.Server {

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
			ctx,
			cfg.HTTP.ShutdownTimeout,
		)
		defer cancel()

		if err := s.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown error", "error", err)
			panic(fmt.Errorf("server shutdown error: %w", err))
		}
	}()

	return s
}
