package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/logger"
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
			fields = append(
				fields,
				zap.String("query", rawQuery),
			)
		}

		ctx := c.Request.Context()

		switch {
		case len(c.Errors) > 0:
			logger.ErrorContext(
				ctx,
				log,
				"http request",
				fields...,
			)

		case status >= 500:
			logger.ErrorContext(
				ctx,
				log,
				"http request",
				fields...,
			)

		case status >= 400:
			logger.WarnContext(
				ctx,
				log,
				"http request",
				fields...,
			)

		default:
			logger.InfoContext(
				ctx,
				log,
				"http request",
				fields...,
			)
		}
	}
}