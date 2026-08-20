package repository

import (
	"context"
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

func (r *OrderRepository) Begin(
	ctx context.Context,
) *gorm.DB {

	return r.db.
		WithContext(ctx).
		Begin()
}

//
// CRUD
//

func (r *OrderRepository) Create(
	ctx context.Context,
	tx *gorm.DB,
	order *models.Order,
) error {

	return tx.
		WithContext(ctx).
		Create(order).
		Error
}

func (r *OrderRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.Order, error) {

	var order models.Order

	err := r.db.
		WithContext(ctx).
		Preload("User").
		Preload("Items").
		Preload("Items.Product").
		Preload("Payment").
		First(
			&order,
			"id = ?",
			id,
		).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		return nil, err
	}

	return &order, nil
}

func (r *OrderRepository) GetByIDTx(
	ctx context.Context,
	tx *gorm.DB,
	id uuid.UUID,
) (*models.Order, error) {

	var order models.Order

	err := tx.
		WithContext(ctx).
		Preload("User").
		Preload("Items").
		Preload("Items.Product").
		Preload("Payment").
		First(
			&order,
			"id = ?",
			id,
		).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		return nil, err
	}

	return &order, nil
}

func (r *OrderRepository) GetAll(
	ctx context.Context,
) ([]models.Order, error) {

	var orders []models.Order

	err := r.db.
		WithContext(ctx).
		Preload("Items").
		Order("created_at DESC").
		Find(&orders).
		Error

	if err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *OrderRepository) GetByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]models.Order, error) {

	var orders []models.Order

	err := r.db.
		WithContext(ctx).
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
	ctx context.Context,
	id uuid.UUID,
	status models.OrderStatus,
) error {

	return r.db.
		WithContext(ctx).
		Model(&models.Order{}).
		Where("id = ?", id).
		Update(
			"status",
			status,
		).
		Error
}

func (r *OrderRepository) UpdateTotalAmount(
	ctx context.Context,
	tx *gorm.DB,
	id uuid.UUID,
	total float64,
) error {

	return tx.
		WithContext(ctx).
		Model(&models.Order{}).
		Where("id = ?", id).
		Update(
			"total_amount",
			total,
		).
		Error
}

// UpdateStatusTx updates order status inside a transaction.
func (r *OrderRepository) UpdateStatusTx(
	ctx context.Context,
	tx *gorm.DB,
	id uuid.UUID,
	status models.OrderStatus,
) error {

	return tx.
		WithContext(ctx).
		Model(&models.Order{}).
		Where("id = ?", id).
		Update(
			"status",
			status,
		).
		Error
}

//
// Delete
//

func (r *OrderRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {

	return r.db.
		WithContext(ctx).
		Delete(
			&models.Order{},
			"id = ?",
			id,
		).
		Error
}
