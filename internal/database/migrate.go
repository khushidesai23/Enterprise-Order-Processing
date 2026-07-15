package database

import (
	"fmt"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

func (d *Database) AutoMigrate() error {

	err := d.DB.AutoMigrate(

		&models.User{},

		&models.Category{},

		&models.Product{},

		&models.Inventory{},

		&models.Order{},

		&models.OrderItem{},

		&models.Payment{},
	)

	if err != nil {
		return fmt.Errorf("migration failed : %w", err)
	}

	return nil
}