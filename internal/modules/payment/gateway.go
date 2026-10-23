package payment

import (
	"context"
	"math"
)

// toPaise converts a rupee amount to the integer paise the gateway expects.
// Rounding (not truncation) matters: 19.99*100 is 1998.9999... in float64.
func toPaise(amount float64) int {
	return int(math.Round(amount * 100))
}

// Gateway Create Order
type GatewayOrderRequest struct {
	Amount   float64
	Currency string
	Receipt  string
}

type GatewayOrderResponse struct {
	// Gateway Order ID
	OrderID string

	Amount float64

	Currency string
}

// Gateway Refund
type GatewayRefundRequest struct {
	// Gateway payment ID (Razorpay pay_...)
	TransactionID string

	Amount float64

	// Our payment ID; sent as the refund receipt for reconciliation.
	Receipt string
}

type GatewayRefundResponse struct {
	// Gateway refund ID (Razorpay rfnd_...)
	RefundID string
}

// Gateway Verification
type GatewayVerificationRequest struct {
	OrderID   string
	PaymentID string
	Signature string
}

// Payment Gateway Interface
type PaymentGateway interface {
	CreateOrder(
		ctx context.Context,
		req GatewayOrderRequest,
	) (*GatewayOrderResponse, error)

	VerifyCheckoutSignature(
		ctx context.Context,
		orderID string,
		paymentID string,
		signature string,
	) error

	VerifyWebhookSignature(
		ctx context.Context,
		body []byte,
		signature string,
	) error

	// RefundPayment refunds a captured payment in full.
	RefundPayment(
		ctx context.Context,
		req GatewayRefundRequest,
	) (*GatewayRefundResponse, error)
}
