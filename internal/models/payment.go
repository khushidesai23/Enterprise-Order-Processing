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

	OrderID uuid.UUID `gorm:"uniqueIndex;not null"`

	Status PaymentStatus `gorm:"size:30;not null"`

	Amount float64 `gorm:"type:numeric(12,2);not null"`

	TransactionID string `gorm:"size:255"`
}