package config

import (
	"log"

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

	LogLevel string
}

var AppConfig Config

func LoadConfig() {

	viper.SetConfigFile(".env")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	AppConfig = Config{
		AppName:    viper.GetString("APP_NAME"),
		AppEnv:     viper.GetString("APP_ENV"),
		AppPort:    viper.GetString("APP_PORT"),
		DBHost:     viper.GetString("DB_HOST"),
		DBPort:     viper.GetString("DB_PORT"),
		DBUser:     viper.GetString("DB_USER"),
		DBPassword: viper.GetString("DB_PASSWORD"),
		DBName:     viper.GetString("DB_NAME"),
		DBSSLMode:  viper.GetString("DB_SSLMODE"),
		LogLevel:   viper.GetString("LOG_LEVEL"),
	}
}