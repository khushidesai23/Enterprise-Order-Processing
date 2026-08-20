package order

import "errors"

var (
	// Order Errors
	ErrOrderNotFound      = errors.New("order not found")
	ErrOrderAlreadyExists = errors.New("order already exists")
	ErrInvalidOrderID     = errors.New("invalid order id")

	// User Errors
	ErrUserNotFound = errors.New("user not found")

	// Product Errors
	ErrProductNotFound = errors.New("product not found")

	// Inventory Errors
	ErrInventoryNotFound = errors.New("inventory not found")

	// Order Item Errors
	ErrOrderItemsRequired = errors.New("at least one order item is required")
	ErrDuplicateProduct   = errors.New("duplicate product found in order")

	// Stock Errors
	ErrInsufficientStock = errors.New("insufficient stock available")

	// Status Errors
	ErrInvalidOrderStatus     = errors.New("invalid order status")
	ErrOrderAlreadyCancelled  = errors.New("order is already cancelled")
	ErrOrderAlreadyCompleted  = errors.New("order is already delivered")
	ErrOrderCannotBeModified  = errors.New("order cannot be modified")
	ErrOrderCannotBeCancelled = errors.New("order cannot be cancelled in the current status")

	// Payment Errors
	ErrPaymentPending = errors.New("payment is pending")
	ErrPaymentFailed  = errors.New("payment failed")

	// Validation Errors
	ErrInvalidOrderRequest = errors.New("invalid order request")
)
