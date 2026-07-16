package product

import "errors"

var (
	// Product Errors
	ErrProductNotFound      = errors.New("product not found")
	ErrProductAlreadyExists = errors.New("product already exists")
	ErrProductSKUExists     = errors.New("product SKU already exists")
	ErrInvalidProductID     = errors.New("invalid product id")

	// Category Errors
	ErrCategoryNotFound = errors.New("category not found")

	// Validation Errors
	ErrInvalidProductName  = errors.New("invalid product name")
	ErrInvalidSKU          = errors.New("invalid SKU")
	ErrInvalidPrice        = errors.New("price must be greater than zero")
	ErrInvalidCategoryID   = errors.New("invalid category id")
	ErrInvalidProductInput = errors.New("invalid product request")

	// Inventory Errors
	ErrInventoryExists = errors.New("inventory already exists")
)