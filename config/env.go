package config

import (
	"fmt"

	"github.com/spf13/viper"
)

func Load() error {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to load .env file: %w", err)
	}

	cfg = &Config{
		App: AppConfig{
			Name: viper.GetString("APP_NAME"),
			Env:  viper.GetString("APP_ENV"),
			Port: viper.GetString("APP_PORT"),
		},

		DB: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			Name:     viper.GetString("DB_NAME"),
			SSLMode:  viper.GetString("DB_SSLMODE"),
		},

		Log: LogConfig{
			Level: viper.GetString("LOG_LEVEL"),
		},
	}

	return validate()
}

func validate() error {

	if cfg.App.Port == "" {
		return fmt.Errorf("APP_PORT is required")
	}

	if cfg.DB.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}

	if cfg.DB.Port == "" {
		return fmt.Errorf("DB_PORT is required")
	}

	if cfg.DB.User == "" {
		return fmt.Errorf("DB_USER is required")
	}

	if cfg.DB.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}

	if cfg.DB.SSLMode == "" {
		cfg.DB.SSLMode = "disable"
	}

	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}

	return nil
}