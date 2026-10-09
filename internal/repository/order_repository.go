package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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

// GetByIDTx loads an order inside a transaction and locks its row
// (FOR UPDATE OF orders) so concurrent status changes are serialized.
// Preloaded associations are read without locks.
func (r *OrderRepository) GetByIDTx(
	ctx context.Context,
	tx *gorm.DB,
	id uuid.UUID,
) (*models.Order, error) {

	var order models.Order

	err := tx.
		WithContext(ctx).
		Clauses(clause.Locking{
			Strength: "UPDATE",
			Table:    clause.Table{Name: clause.CurrentTable},
		}).
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
	limit int,
	offset int,
) ([]models.Order, error) {

	var orders []models.Order

	err := r.db.
		WithContext(ctx).
		Preload("Items").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
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

func (r *OrderRepository) GetByUserIDPage(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
	offset int,
) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.
		WithContext(ctx).
		Preload("Items").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
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
