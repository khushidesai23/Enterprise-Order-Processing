package database

import (
	"fmt"
	"strings"

	"gorm.io/gorm"

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

		&models.PaymentWebhook{},

		&models.Review{},
	)

	if err != nil {
		return fmt.Errorf("migration failed : %w", err)
	}

	// Rows are soft-deleted, so a unique value must be unique among live
	// rows only; otherwise a deleted category name, product SKU or user
	// email can never be used again. AutoMigrate leaves an existing index
	// alone, so databases created before these indexes became partial are
	// upgraded here.
	for _, index := range liveUniqueIndexes {
		if err := d.ensureLiveUniqueIndex(index); err != nil {
			return fmt.Errorf("migration failed : %w", err)
		}
	}

	return nil
}

type liveUniqueIndex struct {
	name   string
	table  string
	column string
}

var liveUniqueIndexes = []liveUniqueIndex{
	{name: "idx_categories_name", table: "categories", column: "name"},
	{name: "idx_products_sku", table: "products", column: "sku"},
	{name: "idx_users_email", table: "users", column: "email"},
}

func (d *Database) ensureLiveUniqueIndex(index liveUniqueIndex) error {
	var definition string

	err := d.DB.Raw(
		"SELECT indexdef FROM pg_indexes WHERE schemaname = current_schema() AND indexname = ?",
		index.name,
	).Scan(&definition).Error
	if err != nil {
		return err
	}

	if strings.Contains(definition, "WHERE") {
		return nil
	}

	return d.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(fmt.Sprintf("DROP INDEX IF EXISTS %s", index.name)).Error; err != nil {
			return err
		}

		return tx.Exec(fmt.Sprintf(
			"CREATE UNIQUE INDEX %s ON %s (%s) WHERE deleted_at IS NULL",
			index.name,
			index.table,
			index.column,
		)).Error
	})
}
