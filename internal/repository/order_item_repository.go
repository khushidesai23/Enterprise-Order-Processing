package repository

import (
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

func (r *OrderItemRepository) Create(
	tx *gorm.DB,
	item *models.OrderItem,
) error {

	return tx.Create(item).Error
}

func (r *OrderItemRepository) CreateMany(
	tx *gorm.DB,
	items []models.OrderItem,
) error {

	return tx.Create(&items).Error
}

func (r *OrderItemRepository) GetByOrderID(
	orderID uuid.UUID,
) ([]models.OrderItem, error) {

	var items []models.OrderItem

	err := r.db.
		Preload("Product").
		Where("order_id = ?", orderID).
		Find(&items).
		Error

	if err != nil {
		return nil, err
	}

	return items, nil
}

func (r *OrderItemRepository) DeleteByOrderID(
	tx *gorm.DB,
	orderID uuid.UUID,
) error {

	return tx.
		Where("order_id = ?", orderID).
		Delete(&models.OrderItem{}).
		Error
}