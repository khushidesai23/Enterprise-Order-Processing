package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{
		db: db,
	}
}

//
// Transaction Support
//

func (r *OrderRepository) Begin() *gorm.DB {
	return r.db.Begin()
}

//
// CRUD
//

func (r *OrderRepository) Create(tx *gorm.DB, order *models.Order) error {
	return tx.Create(order).Error
}

func (r *OrderRepository) GetByID(id uuid.UUID) (*models.Order, error) {

	var order models.Order

	err := r.db.
		Preload("User").
		Preload("Items").
		Preload("Items.Product").
		Preload("Payment").
		First(&order, "id = ?", id).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		return nil, err
	}

	return &order, nil
}

func (r *OrderRepository) GetAll() ([]models.Order, error) {

	var orders []models.Order

	err := r.db.
		Preload("Items").
		Order("created_at DESC").
		Find(&orders).
		Error

	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *OrderRepository) GetByUserID(userID uuid.UUID) ([]models.Order, error) {

	var orders []models.Order

	err := r.db.
		Preload("Items").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&orders).
		Error

	if err != nil {
		return nil, err
	}

	return orders, nil
}

//
// Update
//
// NOTE:
// Never use Save() because Order is loaded with Preload()
// and Save() attempts to persist the whole object graph.
//

func (r *OrderRepository) UpdateStatus(
	id uuid.UUID,
	status models.OrderStatus,
) error {

	return r.db.
		Model(&models.Order{}).
		Where("id = ?", id).
		Update("status", status).
		Error
}

func (r *OrderRepository) UpdateTotalAmount(
	tx *gorm.DB,
	id uuid.UUID,
	total float64,
) error {

	return tx.
		Model(&models.Order{}).
		Where("id = ?", id).
		Update("total_amount", total).
		Error
}

// UpdateStatusTx
func (r *OrderRepository) UpdateStatusTx(
	tx *gorm.DB,
	id uuid.UUID,
	status models.OrderStatus,
) error {

	return tx.
		Model(&models.Order{}).
		Where("id = ?", id).
		Update("status", status).
		Error
}

//
// Delete
//

func (r *OrderRepository) Delete(id uuid.UUID) error {

	return r.db.
		Delete(&models.Order{}, "id = ?", id).
		Error
}