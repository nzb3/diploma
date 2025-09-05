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
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       cfg.IdleTimeout,
		MaxHeaderBytes:    cfg.MaxHeaderBytes,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(
			ctx,
			cfg.ShutdownTimeout,
		)
		defer cancel()

		if err := s.Shutdown(shutdownCtx); err != nil {
			slog.Error("server shutdown error", "error", err)
			panic(fmt.Errorf("server shutdown error: %w", err))
		}
	}()

	return s
}
