package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

type OrderItemRepository struct {
	db *gorm.DB
}

func NewOrderItemRepository(db *gorm.DB) *OrderItemRepository {
	return &OrderItemRepository{
		db: db,
	}
}

// CreateMany creates multiple order items inside a transaction.
func (r *OrderItemRepository) CreateMany(
	ctx context.Context,
	tx *gorm.DB,
	items []models.OrderItem,
) error {

	return tx.
		WithContext(ctx).
		Create(&items).
		Error
}
