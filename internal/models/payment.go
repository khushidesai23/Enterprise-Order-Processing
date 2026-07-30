package models

import "github.com/google/uuid"

type PaymentStatus string

const (
	PaymentPending  PaymentStatus = "PENDING"
	PaymentSuccess  PaymentStatus = "SUCCESS"
	PaymentFailed   PaymentStatus = "FAILED"
	PaymentRefunded PaymentStatus = "REFUNDED"
)

type Payment struct {
	BaseModel

	// Internal Order
	OrderID uuid.UUID `gorm:"uniqueIndex;not null"`

	// Razorpay Order ID
	GatewayOrderID string `gorm:"size:255;uniqueIndex;not null"`

	// Razorpay Payment ID
	TransactionID *string `gorm:"size:255;uniqueIndex"`

	Status PaymentStatus `gorm:"size:30;not null"`

	Amount float64 `gorm:"type:numeric(12,2);not null"`

	Currency string `gorm:"size:10;not null;default:'INR'"`

	Gateway string `gorm:"size:30;not null;default:'RAZORPAY'"`
}