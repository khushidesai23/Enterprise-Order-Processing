package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/metrics"
)

func TestPrometheusSkipsScrapeEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Prometheus())
	router.GET(MetricsPath, func(c *gin.Context) { c.Status(http.StatusOK) })
	router.GET("/api/v1/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	scrape := metrics.HTTPRequestsTotal.WithLabelValues(http.MethodGet, MetricsPath, "200")
	ping := metrics.HTTPRequestsTotal.WithLabelValues(http.MethodGet, "/api/v1/ping", "200")
	scrapeBefore, pingBefore := counterValue(t, scrape), counterValue(t, ping)

	for _, path := range []string{MetricsPath, "/api/v1/ping"} {
		router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}

	assert.Equal(t, scrapeBefore, counterValue(t, scrape))
	assert.Equal(t, pingBefore+1, counterValue(t, ping))
}

func counterValue(t *testing.T, counter prometheus.Counter) float64 {
	t.Helper()
	var metric dto.Metric
	require.NoError(t, counter.Write(&metric))
	return metric.GetCounter().GetValue()
}
