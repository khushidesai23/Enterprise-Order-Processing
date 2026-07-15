package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/khushidesai23/Enterprise-Order-Processing/config"
)

type Database struct {
	DB *gorm.DB
}

func New(cfg *config.Config) (*Database, error) {

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %w", err)
	}

	sqlDB, err := db.DB()

	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)
	sqlDB.SetConnMaxLifetime(2 * time.Hour)

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return &Database{
		DB: db,
	}, nil
}

func (d *Database) Close() error {

	sqlDB, err := d.DB.DB()

	if err != nil {
		return err
	}

	return sqlDB.Close()
}