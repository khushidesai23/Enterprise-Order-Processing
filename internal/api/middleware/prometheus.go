package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/metrics"
)

// MetricsPath is the Prometheus scrape endpoint. Scrapes are excluded from
// HTTP metrics, traces and access logs so they do not skew load-test data.
const MetricsPath = "/metrics"

func Prometheus() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == MetricsPath {
			c.Next()
			return
		}

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
