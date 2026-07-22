package payment

import "context"

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

// Gateway Verification
type GatewayVerificationRequest struct {
	OrderID    string
	PaymentID  string
	Signature  string
}

// Gateway Callback
type GatewayCallback struct {
	OrderID       string
	PaymentID     string
	PaymentStatus string
	Signature     string
}

// Payment Gateway Interface
type PaymentGateway interface {

	CreateOrder(
		ctx context.Context,
		req GatewayOrderRequest,
	) (*GatewayOrderResponse, error)

	VerifySignature(
		ctx context.Context,
		req GatewayVerificationRequest,
	) error
}