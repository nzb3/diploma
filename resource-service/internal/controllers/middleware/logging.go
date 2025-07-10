package middleware

import (
	"log/slog"
	"net/http/httputil"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		if c.Request.Body != nil {
			if dump, err := httputil.DumpRequest(c.Request, true); err == nil {
				slog.Debug("Incoming request", "dump", string(dump))
			}
		}

		c.Next()

		slog.Info("Request processed",
			"path", path,
			"status", c.Writer.Status(),
			"duration", time.Since(start),
			"client_ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		)
	}
}
