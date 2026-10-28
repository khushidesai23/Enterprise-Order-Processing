package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func corsRouter(origins []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(CORS(origins))
	router.POST("/thing", func(c *gin.Context) { c.Status(http.StatusOK) })
	return router
}

func corsRequest(router *gin.Engine, origin string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/thing", nil)
	request.Header.Set("Origin", origin)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestCORSAllowsConfiguredOrigins(t *testing.T) {
	router := corsRouter([]string{"https://shop.example.com"})

	allowed := corsRequest(router, "https://shop.example.com")
	assert.Equal(t, http.StatusOK, allowed.Code)
	assert.Equal(t, "https://shop.example.com", allowed.Header().Get("Access-Control-Allow-Origin"))

	assert.Equal(t, http.StatusForbidden, corsRequest(router, "https://evil.example.com").Code)
}

func TestCORSWithNoOriginsRefusesCrossOriginRequests(t *testing.T) {
	router := corsRouter(nil)

	assert.Equal(t, http.StatusForbidden, corsRequest(router, "https://shop.example.com").Code)
}
