package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

type PaymentRepository struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{
		db: db,
	}
}

// Transaction Support

func (r *PaymentRepository) Begin(
	ctx context.Context,
) *gorm.DB {

	return r.db.
		WithContext(ctx).
		Begin()
}

// Create

func (r *PaymentRepository) Create(
	ctx context.Context,
	tx *gorm.DB,
	payment *models.Payment,
) error {

	return tx.
		WithContext(ctx).
		Create(payment).
		Error
}

// Read

func (r *PaymentRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.Payment, error) {

	var payment models.Payment

	err := r.db.
		WithContext(ctx).
		First(
			&payment,
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

	return &payment, nil
}

func (r *PaymentRepository) GetByIDTx(
	ctx context.Context,
	tx *gorm.DB,
	id uuid.UUID,
) (*models.Payment, error) {

	var payment models.Payment

	err := tx.
		WithContext(ctx).
		First(
			&payment,
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

	return &payment, nil
}

func (r *PaymentRepository) GetByOrderID(
	ctx context.Context,
	orderID uuid.UUID,
) (*models.Payment, error) {

	var payment models.Payment

	err := r.db.
		WithContext(ctx).
		Where(
			"order_id = ?",
			orderID,
		).
		First(&payment).
		Error

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		return nil, err
	}

	return &payment, nil
}

func (r *PaymentRepository) GetByOrderIDTx(
	ctx context.Context,
	tx *gorm.DB,
	orderID uuid.UUID,
) (*models.Payment, error) {

	var payment models.Payment

	err := tx.
		WithContext(ctx).
		Where(
			"order_id = ?",
			orderID,
		).
		First(&payment).
		Error

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		return nil, err
	}

	return &payment, nil
}

func (r *PaymentRepository) GetByGatewayOrderID(
	ctx context.Context,
	gatewayOrderID string,
) (*models.Payment, error) {

	var payment models.Payment

	err := r.db.
		WithContext(ctx).
		Where(
			"gateway_order_id = ?",
			gatewayOrderID,
		).
		First(&payment).
		Error

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		return nil, err
	}

	return &payment, nil
}

func (r *PaymentRepository) GetByGatewayOrderIDTx(
	ctx context.Context,
	tx *gorm.DB,
	gatewayOrderID string,
) (*models.Payment, error) {

	var payment models.Payment

	err := tx.
		WithContext(ctx).
		Where(
			"gateway_order_id = ?",
			gatewayOrderID,
		).
		First(&payment).
		Error

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		return nil, err
	}

	return &payment, nil
}

func (r *PaymentRepository) GetByTransactionID(
	ctx context.Context,
	transactionID string,
) (*models.Payment, error) {

	var payment models.Payment

	err := r.db.
		WithContext(ctx).
		Where(
			"transaction_id = ?",
			transactionID,
		).
		First(&payment).
		Error

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		return nil, err
	}

	return &payment, nil
}

func (r *PaymentRepository) GetAll(
	ctx context.Context,
) ([]models.Payment, error) {

	var payments []models.Payment

	err := r.db.
		WithContext(ctx).
		Order("created_at DESC").
		Find(&payments).
		Error

	if err != nil {
		return nil, err
	}

	return payments, nil
}

// Update

func (r *PaymentRepository) UpdateStatus(
	ctx context.Context,
	tx *gorm.DB,
	id uuid.UUID,
	status models.PaymentStatus,
) error {

	return tx.
		WithContext(ctx).
		Model(&models.Payment{}).
		Where(
			"id = ?",
			id,
		).
		Update(
			"status",
			status,
		).
		Error
}

func (r *PaymentRepository) UpdateTransaction(
	ctx context.Context,
	tx *gorm.DB,
	id uuid.UUID,
	transactionID *string,
) error {

	return tx.
		WithContext(ctx).
		Model(&models.Payment{}).
		Where(
			"id = ?",
			id,
		).
		Update(
			"transaction_id",
			transactionID,
		).
		Error
}

func (r *PaymentRepository) UpdateGatewayOrder(
	ctx context.Context,
	tx *gorm.DB,
	id uuid.UUID,
	gatewayOrderID string,
) error {

	return tx.
		WithContext(ctx).
		Model(&models.Payment{}).
		Where(
			"id = ?",
			id,
		).
		Update(
			"gateway_order_id",
			gatewayOrderID,
		).
		Error
}

// Complete Payment

func (r *PaymentRepository) CompletePayment(
	ctx context.Context,
	tx *gorm.DB,
	id uuid.UUID,
	transactionID *string,
	status models.PaymentStatus,
) error {

	return tx.
		WithContext(ctx).
		Model(&models.Payment{}).
		Where(
			"id = ?",
			id,
		).
		Updates(map[string]interface{}{
			"transaction_id": transactionID,
			"status":         status,
		}).
		Error
}

// Delete

func (r *PaymentRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {

	return r.db.
		WithContext(ctx).
		Delete(
			&models.Payment{},
			"id = ?",
			id,
		).
		Error
}
