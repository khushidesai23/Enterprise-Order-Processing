package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/handlers"
)

func SetupRouter() *gin.Engine {

	router := gin.Default()

	router.GET("/health", handlers.Health)
	router.GET("/ping", handlers.Ping)

	return router
}