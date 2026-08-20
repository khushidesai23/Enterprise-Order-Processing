//go:build e2e

package integration

import (
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/khushidesai23/Enterprise-Order-Processing/config"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/app"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/database"
)

type AppEnv struct {
	Config  *config.Config
	DB      *database.Database
	Router  *gin.Engine
	Gateway *StubPaymentGateway
}

func NewAppEnv(t testing.TB) *AppEnv {
	t.Helper()

	cfg := NewTestConfig()
	db := NewTestDatabase(t)
	CleanupDatabase(t, db)

	gateway := NewStubPaymentGateway(
		cfg.RazorpayKeyID,
		cfg.RazorpayKeySecret,
		cfg.RazorpayWebhookSecret,
	)

	router, err := app.NewRouter(app.RouterOptions{
		Config:         cfg,
		Database:       db,
		Logger:         NewTestLogger(t),
		PaymentGateway: gateway,
	})
	if err != nil {
		t.Fatalf("app.NewRouter() error = %v", err)
	}

	return &AppEnv{
		Config:  cfg,
		DB:      db,
		Router:  router,
		Gateway: gateway,
	}
}
