package payment

import "errors"

var (
	// Payment Errors
	ErrPaymentNotFound         = errors.New("payment not found")
	ErrPaymentAlreadyCompleted = errors.New("payment already completed")
	ErrPaymentPending          = errors.New("payment is still pending")
	ErrPaymentNotRefundable    = errors.New("only captured payments can be refunded")

	// Order Errors
	ErrOrderNotFound              = errors.New("order not found")
	ErrOrderAlreadyPaid           = errors.New("order is already paid")
	ErrOrderCancelled             = errors.New("order has already been cancelled")
	ErrOrderNotEligibleForPayment = errors.New("order is not eligible for payment")

	// Gateway Errors
	ErrGatewayOrderCreation = errors.New("failed to create payment order with gateway")
	ErrGatewayVerification  = errors.New("payment verification failed")
	ErrInvalidSignature     = errors.New("invalid payment signature")
	ErrUnsupportedGateway   = errors.New("unsupported payment gateway")
	ErrGatewayRefund        = errors.New("payment gateway refund failed")

	// Webhook Errors
	ErrDuplicateWebhook    = errors.New("duplicate webhook received")
	ErrWebhookValidation   = errors.New("webhook validation failed")
	ErrUnknownWebhookEvent = errors.New("unknown webhook event")
	ErrMissingWebhookEvent = errors.New("missing webhook event id")

	// Transaction Errors
	ErrTransactionNotFound  = errors.New("transaction not found")
	ErrDuplicateTransaction = errors.New("duplicate transaction")

	// Validation Errors
	ErrInvalidPaymentRequest = errors.New("invalid payment request")
	ErrInvalidPaymentStatus  = errors.New("invalid payment status")
)
