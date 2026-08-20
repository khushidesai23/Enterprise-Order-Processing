package inventory

import "errors"

var (
	// Inventory Errors
	ErrInventoryNotFound      = errors.New("inventory not found")
	ErrInventoryAlreadyExists = errors.New("inventory already exists")
	ErrInvalidInventoryID     = errors.New("invalid inventory id")

	// Product Errors
	ErrProductNotFound = errors.New("product not found")

	// Stock Errors
	ErrInsufficientStock    = errors.New("insufficient available stock")
	ErrInvalidQuantity      = errors.New("quantity must be greater than zero")
	ErrNegativeStock        = errors.New("stock cannot be negative")
	ErrReservedStockExceeds = errors.New("reserved quantity exceeds available stock")
	ErrInsufficientReserved = errors.New("insufficient reserved stock")

	// Validation Errors
	ErrInvalidInventoryInput = errors.New("invalid inventory request")
)
