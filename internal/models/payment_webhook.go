package models

type WebhookStatus string

const (
	WebhookPending   WebhookStatus = "PENDING"
	WebhookProcessed WebhookStatus = "PROCESSED"
	WebhookFailed    WebhookStatus = "FAILED"
)

type PaymentWebhook struct {
	BaseModel

	// Payment Gateway
	Gateway string `gorm:"size:30;not null"`

	// Unique Razorpay Payload ID
	PayloadID string `gorm:"size:255;uniqueIndex;not null"`

	// payment.captured
	// payment.failed
	// refund.created
	Event string `gorm:"size:100;not null"`

	// Merchant Account ID
	AccountID string `gorm:"size:255"`

	// Razorpay Payment ID
	TransactionID string `gorm:"size:255"`

	// Razorpay Order ID
	GatewayOrderID string `gorm:"size:255"`

	Status WebhookStatus `gorm:"size:20;default:'PENDING'"`

	// Complete webhook payload
	RawPayload []byte `gorm:"type:jsonb"`

	// Unix Timestamp
	ProcessedAt *int64
}