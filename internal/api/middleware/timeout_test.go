package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/response"
)

func TestRequestTimeoutReturnsGatewayTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestTimeout(20 * time.Millisecond))
	router.GET("/slow", func(c *gin.Context) {
		// Stands in for a database call blocked past the deadline.
		<-c.Request.Context().Done()
		response.Error(c, http.StatusInternalServerError, c.Request.Context().Err().Error())
	})
	router.GET("/fast", func(c *gin.Context) {
		response.Error(c, http.StatusInternalServerError, "boom")
	})

	slow := httptest.NewRecorder()
	router.ServeHTTP(slow, httptest.NewRequest(http.MethodGet, "/slow", nil))
	fast := httptest.NewRecorder()
	router.ServeHTTP(fast, httptest.NewRequest(http.MethodGet, "/fast", nil))

	assert.Equal(t, http.StatusGatewayTimeout, slow.Code)
	assert.Contains(t, slow.Body.String(), "request timed out")
	assert.Equal(t, http.StatusInternalServerError, fast.Code)
}
