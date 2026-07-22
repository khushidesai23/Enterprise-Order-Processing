package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/khushidesai23/Enterprise-Order-Processing/config"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/handlers"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/middleware"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/routes"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/database"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/category"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/inventory"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/order"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/payment"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/product"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/user"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/repository"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/logger"
)

func main() {

	// Load Configuration
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Logger
	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = log.Sync()
	}()

	// Gin Mode
	switch cfg.AppEnv {
	case "production":
		gin.SetMode(gin.ReleaseMode)

	case "test":
		gin.SetMode(gin.TestMode)

	default:
		gin.SetMode(gin.DebugMode)
	}

	// Database
	db, err := database.New(cfg)
	if err != nil {
		log.Fatal("database connection failed", zap.Error(err))
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Error("failed to close database", zap.Error(err))
		}
	}()

	// Auto Migration
	if err := db.AutoMigrate(); err != nil {
		log.Fatal("database migration failed", zap.Error(err))
	}

	log.Info("database migration completed")

	// Dependency Injection
	healthHandler := handlers.NewHealthHandler(cfg, db)

	userRepository := repository.NewUserRepository(db.DB)
	userService := user.NewService(userRepository)
	userHandler := user.NewHandler(userService)

	productRepository := repository.NewProductRepository(db.DB)
	productService := product.NewService(productRepository)
	productHandler := product.NewHandler(productService)

	categoryRepository := repository.NewCategoryRepository(db.DB)
	categoryService := category.NewService(categoryRepository)
	categoryHandler := category.NewHandler(categoryService)

	inventoryRepository := repository.NewInventoryRepository(db.DB)
	inventoryService := inventory.NewService(inventoryRepository, productRepository)
	inventoryHandler := inventory.NewHandler(inventoryService)

	orderRepository := repository.NewOrderRepository(db.DB)
	orderItemRepository := repository.NewOrderItemRepository(db.DB)
	orderService := order.NewService(orderRepository, orderItemRepository, userRepository, productRepository, inventoryRepository)
	orderHandler := order.NewHandler(orderService)

	paymentRepository := repository.NewPaymentRepository(db.DB)
	paymentWebhookRepository := repository.NewPaymentWebhookRepository(db.DB)

	gateway := payment.NewRazorpayGateway(
		cfg.RazorpayKeyID,
		cfg.RazorpayKeySecret,
		cfg.RazorpayWebhookSecret,
	)

	paymentService := payment.NewService(
		paymentRepository,
		paymentWebhookRepository,
		orderRepository,
		inventoryRepository,
		gateway,
		cfg.RazorpayKeyID,
	)
	paymentHandler := payment.NewHandler(paymentService)

	// Router
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger(log))

	routes.Register(
		router,
		healthHandler,
		userHandler,
		productHandler,
		categoryHandler,
		inventoryHandler,
		orderHandler,
		paymentHandler,
	)

	// HTTP Server
	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start Server
	go func() {

		log.Info(
			"starting server",
			zap.String("application", cfg.AppName),
			zap.String("environment", cfg.AppEnv),
			zap.String("port", cfg.AppPort),
		)

		if err := server.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {

			log.Fatal("server crashed", zap.Error(err))
		}
	}()

	// Graceful Shutdown
	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGTERM,
	)

	<-stop

	log.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server shutdown failed", zap.Error(err))
	}

	log.Info("server stopped gracefully")

	fmt.Println("Application stopped.")
}
