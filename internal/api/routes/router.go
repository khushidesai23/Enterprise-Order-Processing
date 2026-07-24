package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/handlers"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/user"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/product"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/category"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/inventory"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/order"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/payment"

)

func Register(
	router *gin.Engine,
	healthHandler *handlers.HealthHandler,
	userHandler *user.Handler,
	productHandler *product.Handler,
	categoryHandler *category.Handler,
	inventoryHandler *inventory.Handler,
	orderHandler *order.Handler,
	paymentHandler *payment.Handler,
) {

	router.GET("/", healthHandler.Root)

	api := router.Group("/api/v1")

	{
		api.GET("/health", healthHandler.Health)
		api.GET("/ready", healthHandler.Ready)
		api.GET("/ping", healthHandler.Ping)
		api.GET("/version", healthHandler.Version)
		user.RegisterRoutes(api, userHandler)
		product.RegisterRoutes(api, productHandler)
		category.RegisterRoutes(api, categoryHandler)
		inventory.RegisterRoutes(api, inventoryHandler)
		order.RegisterRoutes(api, orderHandler)
		payment.RegisterRoutes(api, paymentHandler)
	}
}