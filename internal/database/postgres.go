package database

import (
	"fmt"
	"time"

	"github.com/khushidesai23/Enterprise-Order-Processing/config"
	applogger "github.com/khushidesai23/Enterprise-Order-Processing/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)
var DB *gorm.DB

func Connect() error {

	cfg := config.Get()

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Name,
		cfg.DB.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})

	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	/*
		Connection Pool

		MaxIdleConns      -> idle connections

		MaxOpenConns      -> total open connections

		ConnMaxLifetime   -> recreate after duration

		ConnMaxIdleTime   -> idle timeout
	*/

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return err
	}

	DB = db

	applogger.L().Info("PostgreSQL connected successfully")

	return nil
}

func GetDB() *gorm.DB {
	return DB
}

func Close() error {

	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()

	if err != nil {
		return err
	}

	applogger.L().Info("Closing PostgreSQL connection")

	return sqlDB.Close()
}