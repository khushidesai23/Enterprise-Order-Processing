package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/handlers"
)

func SetupRouter() *gin.Engine {

	router := gin.New()

	router.Use(
		gin.Logger(),
		gin.Recovery(),
	)

	router.GET("/", func(c *gin.Context) {

		c.JSON(http.StatusOK, gin.H{
			"service": "Enterprise Order Processing",
			"version": "v1",
			"status":  "running",
		})

	})

	api := router.Group("/api/v1")

	{
		api.GET("/health", handlers.Health)
		api.GET("/ping", handlers.Ping)
	}

	return router
}