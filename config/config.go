package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppName string
	AppEnv  string
	AppPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	RazorpayKeyID         string
	RazorpayKeySecret     string
	RazorpayWebhookSecret string

	JWTSecret     string
	JWTExpiration time.Duration

	LogLevel string

	OTelServiceName      string
	OTelServiceVersion   string
	OTelEnvironment      string
	OTelExporterEndpoint string

	KafkaBrokers    []string
	KafkaCDCGroupID string
	KafkaCDCTopic   string
	KafkaMinBytes   int
	KafkaMaxBytes   int
	KafkaMaxWait    time.Duration
}

func Load() (*Config, error) {
	if err := loadDotEnv(".env"); err != nil {
		return nil, err
	}

	setDefaults()

	viper.AutomaticEnv()

	cfg := &Config{
		AppName: viper.GetString("APP_NAME"),
		AppEnv:  viper.GetString("APP_ENV"),
		AppPort: viper.GetString("APP_PORT"),

		DBHost:     viper.GetString("DB_HOST"),
		DBPort:     viper.GetString("DB_PORT"),
		DBUser:     viper.GetString("DB_USER"),
		DBPassword: viper.GetString("DB_PASSWORD"),
		DBName:     viper.GetString("DB_NAME"),
		DBSSLMode:  viper.GetString("DB_SSLMODE"),

		RazorpayKeyID:         viper.GetString("RAZORPAY_KEY_ID"),
		RazorpayKeySecret:     viper.GetString("RAZORPAY_KEY_SECRET"),
		RazorpayWebhookSecret: viper.GetString("RAZORPAY_WEBHOOK_SECRET"),

		JWTSecret: viper.GetString("JWT_SECRET"),
		JWTExpiration: mustParseDuration(
			viper.GetString("JWT_EXPIRATION"),
		),

		LogLevel: viper.GetString("LOG_LEVEL"),

		OTelServiceName:      viper.GetString("OTEL_SERVICE_NAME"),
		OTelServiceVersion:   viper.GetString("OTEL_SERVICE_VERSION"),
		OTelEnvironment:      viper.GetString("OTEL_ENVIRONMENT"),
		OTelExporterEndpoint: viper.GetString("OTEL_EXPORTER_OTLP_ENDPOINT"),

		KafkaBrokers: strings.Split(viper.GetString("KAFKA_BROKERS"), ","),
		KafkaCDCGroupID: viper.GetString("KAFKA_CDC_GROUP_ID"),
		KafkaCDCTopic:   viper.GetString("KAFKA_CDC_TOPIC"),
		KafkaMinBytes:   viper.GetInt("KAFKA_MIN_BYTES"),
		KafkaMaxBytes:   viper.GetInt("KAFKA_MAX_BYTES"),
		KafkaMaxWait: mustParseDuration(
			viper.GetString("KAFKA_MAX_WAIT"),
		),
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func setDefaults() {
	viper.SetDefault("APP_NAME", "Enterprise Order Processing")
	viper.SetDefault("APP_ENV", "development")
	viper.SetDefault("APP_PORT", "8080")

	viper.SetDefault("DB_HOST", "localhost")
	viper.SetDefault("DB_PORT", "5432")
	viper.SetDefault("DB_USER", "postgres")
	viper.SetDefault("DB_PASSWORD", "postgres")
	viper.SetDefault("DB_NAME", "order_processing")
	viper.SetDefault("DB_SSLMODE", "disable")

	viper.SetDefault("RAZORPAY_KEY_ID", "")
	viper.SetDefault("RAZORPAY_KEY_SECRET", "")
	viper.SetDefault("RAZORPAY_WEBHOOK_SECRET", "")

	viper.SetDefault("JWT_SECRET", "")
	viper.SetDefault("JWT_EXPIRATION", 24*time.Hour)

	viper.SetDefault("LOG_LEVEL", "debug")

	viper.SetDefault(
		"OTEL_SERVICE_NAME",
		"enterprise-order-processing",
	)

	viper.SetDefault(
		"OTEL_SERVICE_VERSION",
		"1.0.0",
	)

	viper.SetDefault(
		"OTEL_ENVIRONMENT",
		"development",
	)

	viper.SetDefault(
		"OTEL_EXPORTER_OTLP_ENDPOINT",
		"localhost:4317",
	)

	viper.SetDefault("KAFKA_BROKERS", []string{"localhost:9092"})
	viper.SetDefault(
		"KAFKA_CDC_GROUP_ID",
		"order-processing-cdc-consumer",
	)
	viper.SetDefault(
		"KAFKA_CDC_TOPIC",
		"order-processing.public.categories",
	)
	viper.SetDefault("KAFKA_MIN_BYTES", 1)
	viper.SetDefault("KAFKA_MAX_BYTES", 10e6)
	viper.SetDefault("KAFKA_MAX_WAIT", "1s")
}

func (c *Config) Validate() error {
	if strings.TrimSpace(c.AppName) == "" {
		return errors.New("APP_NAME is required")
	}
	if strings.TrimSpace(c.AppEnv) == "" {
		return errors.New("APP_ENV is required")
	}
	if strings.TrimSpace(c.AppPort) == "" {
		return errors.New("APP_PORT is required")
	}
	if strings.TrimSpace(c.DBHost) == "" {
		return errors.New("DB_HOST is required")
	}
	if strings.TrimSpace(c.DBPort) == "" {
		return errors.New("DB_PORT is required")
	}
	if strings.TrimSpace(c.DBUser) == "" {
		return errors.New("DB_USER is required")
	}
	if strings.TrimSpace(c.DBName) == "" {
		return errors.New("DB_NAME is required")
	}
	if strings.TrimSpace(c.DBSSLMode) == "" {
		return errors.New("DB_SSLMODE is required")
	}
	if strings.TrimSpace(c.RazorpayKeyID) == "" {
		return errors.New("RAZORPAY_KEY_ID is required")
	}
	if strings.TrimSpace(c.RazorpayKeySecret) == "" {
		return errors.New("RAZORPAY_KEY_SECRET is required")
	}
	if strings.TrimSpace(c.RazorpayWebhookSecret) == "" {
		return errors.New("RAZORPAY_WEBHOOK_SECRET is required")
	}

	if strings.TrimSpace(c.JWTSecret) == "" {
		return errors.New("JWT_SECRET is required")
	}
	if c.JWTExpiration <= 0 {
		return errors.New("JWT_EXPIRATION is required")
	}

	if strings.TrimSpace(c.LogLevel) == "" {
		return errors.New("LOG_LEVEL is required")
	}

	if len(c.KafkaBrokers) == 0 {
		return errors.New("KAFKA_BROKERS is required")
	}

	if strings.TrimSpace(c.KafkaCDCGroupID) == "" {
		return errors.New("KAFKA_CDC_GROUP_ID is required")
	}

	if strings.TrimSpace(c.KafkaCDCTopic) == "" {
		return errors.New("KAFKA_CDC_TOPIC is required")
	}

	if c.KafkaMinBytes <= 0 {
		return errors.New("KAFKA_MIN_BYTES must be greater than 0")
	}

	if c.KafkaMaxBytes <= 0 {
		return errors.New("KAFKA_MAX_BYTES must be greater than 0")
	}

	if c.KafkaMaxWait <= 0 {
		return errors.New("KAFKA_MAX_WAIT must be greater than 0")
	}

	return nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		c.DBHost,
		c.DBPort,
		c.DBUser,
		c.DBPassword,
		c.DBName,
		c.DBSSLMode,
	)
}

func loadDotEnv(path string) error {
	absPath := path

	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		absPath = filepath.Join(cwd, path)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if info.IsDir() {
		return fmt.Errorf("%s is a directory", absPath)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)

		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}

	return scanner.Err()
}

func mustParseDuration(value string) time.Duration {
	d, err := time.ParseDuration(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return d
}
