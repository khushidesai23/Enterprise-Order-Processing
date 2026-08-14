package repository

import (
	"context"
	"errors"
	"time"

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

// Create creates a new payment webhook record inside a transaction.
func (r *PaymentWebhookRepository) Create(
	ctx context.Context,
	tx *gorm.DB,
	webhook *models.PaymentWebhook,
) error {

	return tx.
		WithContext(ctx).
		Create(webhook).
		Error
}

// GetByPayloadID returns a webhook by its payload ID.
func (r *PaymentWebhookRepository) GetByPayloadID(
	ctx context.Context,
	payloadID string,
) (*models.PaymentWebhook, error) {

	var webhook models.PaymentWebhook

	err := r.db.
		WithContext(ctx).
		Where(
			"payload_id = ?",
			payloadID,
		).
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

// MarkProcessed marks a webhook as successfully processed.
func (r *PaymentWebhookRepository) MarkProcessed(
	ctx context.Context,
	tx *gorm.DB,
	payloadID string,
	processedAt time.Time,
) error {

	return tx.
		WithContext(ctx).
		Model(&models.PaymentWebhook{}).
		Where(
			"payload_id = ?",
			payloadID,
		).
		Updates(map[string]interface{}{
			"status":       models.WebhookProcessed,
			"processed_at": processedAt,
		}).
		Error
}

// MarkFailed marks a webhook as failed.
func (r *PaymentWebhookRepository) MarkFailed(
	ctx context.Context,
	tx *gorm.DB,
	payloadID string,
	processedAt time.Time,
) error {

	return tx.
		WithContext(ctx).
		Model(&models.PaymentWebhook{}).
		Where(
			"payload_id = ?",
			payloadID,
		).
		Updates(map[string]interface{}{
			"status":       models.WebhookFailed,
			"processed_at": processedAt,
		}).
		Error
}