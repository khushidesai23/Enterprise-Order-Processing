package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/handlers"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/user"
)

func Register(
	router *gin.Engine,
	healthHandler *handlers.HealthHandler,
	userHandler *user.Handler,
) {

	router.GET("/", healthHandler.Root)

	api := router.Group("/api/v1")

	{
		api.GET("/health", healthHandler.Health)

		api.GET("/ready", healthHandler.Ready)

		api.GET("/ping", healthHandler.Ping)

		api.GET("/version", healthHandler.Version)

		user.RegisterRoutes(api, userHandler)
	}
}