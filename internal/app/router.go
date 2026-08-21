package app

import (
	"errors"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"

	"github.com/khushidesai23/Enterprise-Order-Processing/config"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/handlers"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/middleware"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/routes"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/database"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/auth"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/category"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/inventory"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/order"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/payment"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/product"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/user"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/repository"
)

type RouterOptions struct {
	Config         *config.Config
	Database       *database.Database
	Logger         *zap.Logger
	PaymentGateway payment.PaymentGateway
}

func NewRouter(options RouterOptions) (*gin.Engine, error) {
	if options.Config == nil {
		return nil, errors.New("config is required")
	}

	if options.Database == nil {
		return nil, errors.New("database is required")
	}

	if options.Logger == nil {
		return nil, errors.New("logger is required")
	}

	switch options.Config.AppEnv {
	case "production":
		gin.SetMode(gin.ReleaseMode)
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		gin.SetMode(gin.DebugMode)
	}

	jwtManager := auth.NewJWTManager(
		options.Config.JWTSecret,
		options.Config.JWTExpiration,
	)

	healthHandler := handlers.NewHealthHandler(
		options.Config,
		options.Database,
	)

	userRepository := repository.NewUserRepository(options.Database.DB)
	userService := user.NewService(userRepository, options.Logger)
	userHandler := user.NewHandler(userService)

	authService := auth.NewService(
		userRepository,
		jwtManager,
		options.Config,
		options.Logger,
	)
	authHandler := auth.NewHandler(authService)

	productRepository := repository.NewProductRepository(options.Database.DB)
	productService := product.NewService(productRepository, options.Logger)
	productHandler := product.NewHandler(productService)

	categoryRepository := repository.NewCategoryRepository(options.Database.DB)
	categoryService := category.NewService(categoryRepository, options.Logger)
	categoryHandler := category.NewHandler(categoryService)

	inventoryRepository := repository.NewInventoryRepository(options.Database.DB)
	inventoryService := inventory.NewService(
		inventoryRepository,
		productRepository,
		options.Logger,
	)
	inventoryHandler := inventory.NewHandler(inventoryService)

	orderRepository := repository.NewOrderRepository(options.Database.DB)
	orderItemRepository := repository.NewOrderItemRepository(options.Database.DB)
	orderService := order.NewService(
		orderRepository,
		orderItemRepository,
		userRepository,
		productRepository,
		inventoryRepository,
		options.Logger,
	)
	orderHandler := order.NewHandler(orderService)

	paymentRepository := repository.NewPaymentRepository(options.Database.DB)
	paymentWebhookRepository := repository.NewPaymentWebhookRepository(options.Database.DB)

	gateway := options.PaymentGateway
	if gateway == nil {
		gateway = payment.NewRazorpayGateway(
			options.Config.RazorpayKeyID,
			options.Config.RazorpayKeySecret,
			options.Config.RazorpayWebhookSecret,
		)
	}

	paymentService := payment.NewService(
		paymentRepository,
		paymentWebhookRepository,
		orderService,
		gateway,
		options.Config.RazorpayKeyID,
		options.Logger,
	)
	paymentHandler := payment.NewHandler(paymentService)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(otelgin.Middleware(options.Config.OTelServiceName))
	router.Use(middleware.Prometheus())
	router.Use(middleware.RequestLogger(options.Logger))
	router.Use(middleware.CORS())

	routes.Register(
		router,
		healthHandler,
		userHandler,
		productHandler,
		categoryHandler,
		inventoryHandler,
		orderHandler,
		paymentHandler,
		authHandler,
		jwtManager,
	)

	return router, nil
}
