package category

import "errors"

var (
	// Category Errors
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryAlreadyExists = errors.New("category already exists")
	ErrCategoryNameExists    = errors.New("category name already exists")
	ErrInvalidCategoryID     = errors.New("invalid category id")

	// Validation Errors
	ErrInvalidCategoryName  = errors.New("invalid category name")
	ErrInvalidCategoryInput = errors.New("invalid category request")

	// Business Errors
	ErrCategoryInUse = errors.New("category is associated with one or more products")
)