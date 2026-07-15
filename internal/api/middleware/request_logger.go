package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func RequestLogger(log *zap.Logger) gin.HandlerFunc {
	if log == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("client_ip", c.ClientIP()),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.Int("body_size", c.Writer.Size()),
		}

		if rawQuery != "" {
			fields = append(fields, zap.String("query", rawQuery))
		}

		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
			log.Error("http request", fields...)
			return
		}

		switch {
		case status >= 500:
			log.Error("http request", fields...)
		case status >= 400:
			log.Warn("http request", fields...)
		default:
			log.Info("http request", fields...)
		}
	}
}