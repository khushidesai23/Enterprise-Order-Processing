package payment

import (
	"github.com/google/uuid"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

// Create Payment
type CreatePaymentRequest struct {
	OrderID uuid.UUID `json:"order_id" binding:"required"`
}

// Razorpay Checkout Response
type CheckoutResponse struct {
	PaymentID uuid.UUID `json:"payment_id"`

	OrderID uuid.UUID `json:"order_id"`

	Amount float64 `json:"amount"`

	Currency string `json:"currency"`

	KeyID string `json:"key_id"`

	Gateway string `json:"gateway"`

	GatewayOrderID string `json:"gateway_order_id"`
}

// Payment Response
type PaymentResponse struct {
	ID uuid.UUID `json:"id"`

	OrderID uuid.UUID `json:"order_id"`

	GatewayOrderID string `json:"gateway_order_id"`

	TransactionID string `json:"transaction_id"`

	Status models.PaymentStatus `json:"status"`

	Amount float64 `json:"amount"`

	Currency string `json:"currency"`

	Gateway string `json:"gateway"`

	CreatedAt string `json:"created_at"`
}

// Payment List Response
type PaymentListResponse struct {
	ID uuid.UUID `json:"id"`

	OrderID uuid.UUID `json:"order_id"`

	Status models.PaymentStatus `json:"status"`

	Amount float64 `json:"amount"`

	Currency string `json:"currency"`

	Gateway string `json:"gateway"`
}

// Razorpay Webhook
type WebhookRequest struct {

	// Razorpay Payment ID
	PaymentID string `json:"razorpay_payment_id"`

	// Razorpay Order ID
	OrderID string `json:"razorpay_order_id"`

	// Signature sent by Razorpay Checkout
	Signature string `json:"razorpay_signature"`

	// Webhook Event
	Event string `json:"event"`
}

// Internal Webhook Processing
type PaymentWebhook struct {
	GatewayOrderID string

	TransactionID string

	Status models.PaymentStatus

	Signature string
}

// Payment Summary
type PaymentSummary struct {
	TotalPayments int64 `json:"total_payments"`

	SuccessfulPayments int64 `json:"successful_payments"`

	FailedPayments int64 `json:"failed_payments"`

	PendingPayments int64 `json:"pending_payments"`

	RefundedPayments int64 `json:"refunded_payments"`

	Revenue float64 `json:"revenue"`
}

// Process Webhook Request
type ProcessWebhookRequest struct {
	Body []byte
	Signature string
	EventID string
}