package integration

import (
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	"github.com/khushidesai23/Enterprise-Order-Processing/config"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/database"
)

const enableDBTestsEnv = "ENABLE_DB_TESTS"

func NewTestConfig() *config.Config {
	return &config.Config{
		AppName:               "Enterprise Order Processing Test",
		AppEnv:                "test",
		AppPort:               "8080",
		DBHost:                envOrDefault("TEST_DB_HOST", "localhost"),
		DBPort:                envOrDefault("TEST_DB_PORT", "5433"),
		DBUser:                envOrDefault("TEST_DB_USER", "postgres"),
		DBPassword:            envOrDefault("TEST_DB_PASSWORD", "postgres"),
		DBName:                envOrDefault("TEST_DB_NAME", "order_processing_test"),
		DBSSLMode:             envOrDefault("TEST_DB_SSLMODE", "disable"),
		RazorpayKeyID:         envOrDefault("TEST_RAZORPAY_KEY_ID", "rzp_test_key"),
		RazorpayKeySecret:     envOrDefault("TEST_RAZORPAY_KEY_SECRET", "rzp_test_secret"),
		RazorpayWebhookSecret: envOrDefault("TEST_RAZORPAY_WEBHOOK_SECRET", "rzp_test_webhook_secret"),
		JWTSecret:             envOrDefault("TEST_JWT_SECRET", "integration-jwt-secret"),
		JWTExpiration:         time.Hour,
		LogLevel:              "error",
		OTelServiceName:       "enterprise-order-processing-test",
		OTelServiceVersion:    "test",
		OTelEnvironment:       "test",
		OTelExporterEndpoint:  envOrDefault("TEST_OTEL_EXPORTER_ENDPOINT", "localhost:4317"),
	}
}

func NewTestDatabase(t testing.TB) *database.Database {
	t.Helper()

	requireDBEnabled(t)

	cfg := NewTestConfig()

	ensureDatabaseExists(t, cfg)

	db, err := database.New(cfg)
	if err != nil {
		t.Fatalf("database.New() error = %v", err)
	}

	if err := db.AutoMigrate(); err != nil {
		_ = db.Close()
		t.Fatalf("db.AutoMigrate() error = %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("db.Close() error = %v", err)
		}
	})

	return db
}

func CleanupDatabase(t testing.TB, db *database.Database) {
	t.Helper()

	statement := strings.Join([]string{
		"TRUNCATE TABLE payment_webhooks, payments, order_items, orders, inventories, products, categories, users RESTART IDENTITY CASCADE",
	}, ";")

	if err := db.DB.Exec(statement).Error; err != nil {
		t.Fatalf("cleanup database error = %v", err)
	}
}

func NewTestLogger(t testing.TB) *zap.Logger {
	t.Helper()
	return zap.NewNop()
}

func ensureDatabaseExists(t testing.TB, cfg *config.Config) {
	t.Helper()

	adminDSN := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s TimeZone=UTC",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPassword,
		cfg.DBSSLMode,
	)

	adminDB, err := sql.Open("pgx", adminDSN)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer adminDB.Close()

	if err := adminDB.Ping(); err != nil {
		t.Fatalf("adminDB.Ping() error = %v", err)
	}

	var exists bool
	row := adminDB.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		cfg.DBName,
	)
	if err := row.Scan(&exists); err != nil {
		t.Fatalf("checking test database existence failed: %v", err)
	}

	if exists {
		return
	}

	query := fmt.Sprintf("CREATE DATABASE %s", quoteIdentifier(cfg.DBName))
	if _, err := adminDB.Exec(query); err != nil {
		t.Fatalf("creating test database failed: %v", err)
	}
}

func requireDBEnabled(t testing.TB) {
	t.Helper()

	if os.Getenv(enableDBTestsEnv) == "" {
		t.Skipf("set %s=1 to run database-backed integration tests", enableDBTestsEnv)
	}
}

func envOrDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func quoteIdentifier(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}
