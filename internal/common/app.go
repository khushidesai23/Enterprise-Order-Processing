package common

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/khushidesai23/Enterprise-Order-Processing/config"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/routes"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/database"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/logger"
)

type Application struct {
	server *http.Server
}

func NewApplication() *Application {

	cfg := config.Get()

	router := routes.SetupRouter()

	server := &http.Server{
		Addr:              ":" + cfg.App.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &Application{
		server: server,
	}
}

func (a *Application) Start() error {

	logger.L().Info("Starting HTTP server")

	go func() {

		if err := a.server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {

			logger.L().Fatal(err.Error())
		}

	}()

	logger.L().Info("Server started successfully")

	stop := make(chan os.Signal, 1)

	signal.Notify(
		stop,
		os.Interrupt,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	<-stop

	logger.L().Info("Shutdown signal received")

	return a.Shutdown()
}

func (a *Application) Shutdown() error {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	defer cancel()

	if err := database.Close(); err != nil {

		logger.L().Error(err.Error())
	}

	logger.L().Info("Stopping HTTP server")

	return a.server.Shutdown(ctx)
}