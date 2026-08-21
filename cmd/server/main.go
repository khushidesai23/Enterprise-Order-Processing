// @title Enterprise Order Processing API
// @version 1.0
// @description Enterprise Order Processing & Payment Platform
// @host localhost:8080
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

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
	_ "github.com/khushidesai23/Enterprise-Order-Processing/docs"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/app"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/database"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/logger"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/metrics"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/telemetry"
)

func main() {

	ctx := context.Background()

	// Load Configuration
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// Register Prometheus metrics
	metrics.Register()

	// Initialize OpenTelemetry
	otelShutdown, err := telemetry.InitTracer(
		ctx,
		cfg.OTelServiceName,
		cfg.OTelServiceVersion,
		cfg.OTelEnvironment,
	)
	if err != nil {
		panic(err)
	}

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			10*time.Second,
		)
		defer cancel()

		if err := otelShutdown(shutdownCtx); err != nil {
			fmt.Printf(
				"failed to shutdown OpenTelemetry: %v\n",
				err,
			)
		}
	}()

	// Logger
	log := logger.New()
	if log == nil {
		panic("failed to initialize logger")
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

	router, err := app.NewRouter(app.RouterOptions{
		Config:   cfg,
		Database: db,
		Logger:   log,
	})
	if err != nil {
		log.Fatal("router initialization failed", zap.Error(err))
	}

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
