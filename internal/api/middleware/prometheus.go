package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/metrics"
)

func Prometheus() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		method := c.Request.Method
		path := c.FullPath()

		if path == "" {
			path = "unknown"
		}

		metrics.HTTPRequestsInFlight.
			WithLabelValues(method, path).
			Inc()

		defer metrics.HTTPRequestsInFlight.
			WithLabelValues(method, path).
			Dec()

		c.Next()

		status := strconv.Itoa(c.Writer.Status())

		duration := time.Since(start).Seconds()

		metrics.HTTPRequestsTotal.
			WithLabelValues(
				method,
				path,
				status,
			).
			Inc()

		metrics.HTTPRequestDuration.
			WithLabelValues(
				method,
				path,
				status,
			).
			Observe(duration)
	}
}
