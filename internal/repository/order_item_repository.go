package repository

import (
	"context"

	"github.com/google/uuid"
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

// Create creates a single order item inside a transaction.
func (r *OrderItemRepository) Create(
	ctx context.Context,
	tx *gorm.DB,
	item *models.OrderItem,
) error {

	return tx.
		WithContext(ctx).
		Create(item).
		Error
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

// GetByOrderID returns all items belonging to an order.
func (r *OrderItemRepository) GetByOrderID(
	ctx context.Context,
	orderID uuid.UUID,
) ([]models.OrderItem, error) {

	var items []models.OrderItem

	err := r.db.
		WithContext(ctx).
		Preload("Product").
		Where(
			"order_id = ?",
			orderID,
		).
		Find(&items).
		Error

	if err != nil {
		return nil, err
	}

	return items, nil
}

// DeleteByOrderID deletes all items belonging to an order
// inside a transaction.
func (r *OrderItemRepository) DeleteByOrderID(
	ctx context.Context,
	tx *gorm.DB,
	orderID uuid.UUID,
) error {

	return tx.
		WithContext(ctx).
		Where(
			"order_id = ?",
			orderID,
		).
		Delete(&models.OrderItem{}).
		Error
}