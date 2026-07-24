package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

type PaymentWebhookRepository struct {
	db *gorm.DB
}

func NewPaymentWebhookRepository(
	db *gorm.DB,
) *PaymentWebhookRepository {

	return &PaymentWebhookRepository{
		db: db,
	}
}

// Create
func (r *PaymentWebhookRepository) Create(
	tx *gorm.DB,
	webhook *models.PaymentWebhook,
) error {

	return tx.Create(webhook).Error
}

// Get
func (r *PaymentWebhookRepository) GetByPayloadID(
	payloadID string,
) (*models.PaymentWebhook, error) {

	var webhook models.PaymentWebhook

	err := r.db.
		Where("payload_id = ?", payloadID).
		First(&webhook).
		Error

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		return nil, err
	}

	return &webhook, nil
}

// Mark Processed
func (r *PaymentWebhookRepository) MarkProcessed(
	tx *gorm.DB,
	payloadID string,
	processedAt int64,
) error {

	return tx.
		Model(&models.PaymentWebhook{}).
		Where("payload_id = ?", payloadID).
		Updates(map[string]interface{}{
			"status":       models.WebhookProcessed,
			"processed_at": processedAt,
		}).
		Error
}

// Mark Failed
func (r *PaymentWebhookRepository) MarkFailed(
	tx *gorm.DB,
	payloadID string,
	processedAt int64,
) error {

	return tx.
		Model(&models.PaymentWebhook{}).
		Where("payload_id = ?", payloadID).
		Updates(map[string]interface{}{
			"status":       models.WebhookFailed,
			"processed_at": processedAt,
		}).
		Error
}