package repository

import (
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
func (r *PaymentRepository) Begin() *gorm.DB {
	return r.db.Begin()
}

// Create
func (r *PaymentRepository) Create(
	tx *gorm.DB,
	payment *models.Payment,
) error {
	return tx.Create(payment).Error
}

// Read
func (r *PaymentRepository) GetByID(
	id uuid.UUID,
) (*models.Payment, error) {

	var payment models.Payment

	err := r.db.
		First(&payment, "id = ?", id).
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
	tx *gorm.DB,
	id uuid.UUID,
) (*models.Payment, error) {

	var payment models.Payment

	err := tx.
		First(&payment, "id = ?", id).
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
	orderID uuid.UUID,
) (*models.Payment, error) {

	var payment models.Payment

	err := r.db.
		Where("order_id = ?", orderID).
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
	tx *gorm.DB,
	orderID uuid.UUID,
) (*models.Payment, error) {

	var payment models.Payment

	err := tx.
		Where("order_id = ?", orderID).
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
	gatewayOrderID string,
) (*models.Payment, error) {

	var payment models.Payment

	err := r.db.
		Where("gateway_order_id = ?", gatewayOrderID).
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
	tx *gorm.DB,
	gatewayOrderID string,
) (*models.Payment, error) {

	var payment models.Payment

	err := tx.
		Where("gateway_order_id = ?", gatewayOrderID).
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
	transactionID string,
) (*models.Payment, error) {

	var payment models.Payment

	err := r.db.
		Where("transaction_id = ?", transactionID).
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

func (r *PaymentRepository) GetAll() ([]models.Payment, error) {

	var payments []models.Payment

	err := r.db.
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
	tx *gorm.DB,
	id uuid.UUID,
	status models.PaymentStatus,
) error {

	return tx.
		Model(&models.Payment{}).
		Where("id = ?", id).
		Update("status", status).
		Error
}

func (r *PaymentRepository) UpdateTransaction(
	tx *gorm.DB,
	id uuid.UUID,
	transactionID string,
) error {

	return tx.
		Model(&models.Payment{}).
		Where("id = ?", id).
		Update("transaction_id", transactionID).
		Error
}

func (r *PaymentRepository) UpdateGatewayOrder(
	tx *gorm.DB,
	id uuid.UUID,
	gatewayOrderID string,
) error {

	return tx.
		Model(&models.Payment{}).
		Where("id = ?", id).
		Update("gateway_order_id", gatewayOrderID).
		Error
}

// Complete Payment
func (r *PaymentRepository) CompletePayment(
	tx *gorm.DB,
	id uuid.UUID,
	transactionID string,
	status models.PaymentStatus,
) error {

	return tx.
		Model(&models.Payment{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"transaction_id": transactionID,
			"status":         status,
		}).
		Error
}

// Delete
func (r *PaymentRepository) Delete(
	id uuid.UUID,
) error {

	return r.db.
		Delete(&models.Payment{}, "id = ?", id).
		Error
}
