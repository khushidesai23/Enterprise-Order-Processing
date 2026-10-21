package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/handlers"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/middleware"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/auth"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/category"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/inventory"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/order"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/payment"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/product"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/user"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	authHandler *auth.Handler,
	jwtManager *auth.JWTManager,
) {

	router.GET("/", healthHandler.Root)

	router.GET(
		"/swagger/*any",
		ginSwagger.WrapHandler(swaggerFiles.Handler),
	)

	router.GET(
		middleware.MetricsPath,
		gin.WrapH(promhttp.Handler()),
	)

	authMiddleware := middleware.AuthMiddleware(jwtManager)
	adminOnly := middleware.RequireRole(models.RoleAdmin)

	api := router.Group("/api/v1")

	{
		// Health
		api.GET("/health", healthHandler.Health)
		api.GET("/ready", healthHandler.Ready)
		api.GET("/ping", healthHandler.Ping)
		api.GET("/version", healthHandler.Version)

		// Modules
		auth.RegisterRoutes(
			api,
			authHandler,
			authMiddleware,
		)

		user.RegisterRoutes(
			api,
			userHandler,
			authMiddleware,
			adminOnly,
		)

		category.RegisterRoutes(
			api,
			categoryHandler,
			authMiddleware,
			adminOnly,
		)

		product.RegisterRoutes(
			api,
			productHandler,
			authMiddleware,
			adminOnly,
		)

		inventory.RegisterRoutes(
			api,
			inventoryHandler,
			authMiddleware,
			adminOnly,
		)

		order.RegisterRoutes(
			api,
			orderHandler,
			authMiddleware,
			adminOnly,
		)

		payment.RegisterRoutes(
			api,
			paymentHandler,
			authMiddleware,
			adminOnly,
		)
	}
}
