package product

import (
	"github.com/google/uuid"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

type CreateProductRequest struct {
	Name        string    `json:"name" binding:"required,min=2,max=255"`
	Description string    `json:"description" binding:"max=1000"`
	SKU         string    `json:"sku" binding:"required,min=2,max=100"`
	Price       float64   `json:"price" binding:"required,gt=0"`
	CategoryID  uuid.UUID `json:"category_id" binding:"required"`
}

type UpdateProductRequest struct {
	Name        string    `json:"name" binding:"required,min=2,max=255"`
	Description string    `json:"description" binding:"max=1000"`
	SKU         string    `json:"sku" binding:"required,min=2,max=100"`
	Price       float64   `json:"price" binding:"required,gt=0"`
	CategoryID  uuid.UUID `json:"category_id" binding:"required"`
}

type ProductResponse struct {
	ID          uuid.UUID        `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	SKU         string           `json:"sku"`
	Price       float64          `json:"price"`
	Category    CategoryResponse `json:"category"`
	Inventory   InventorySummary `json:"inventory,omitempty"`
}

type ProductListResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	SKU        string    `json:"sku"`
	Price      float64   `json:"price"`
	CategoryID uuid.UUID `json:"category_id"`
	Category   string    `json:"category"`
	Available  int       `json:"available_quantity"`
	Reserved   int       `json:"reserved_quantity"`
}

type CategoryResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type InventorySummary struct {
	AvailableQuantity int `json:"available_quantity"`
	ReservedQuantity  int `json:"reserved_quantity"`
}

type ProductQuery struct {
	Page       int
	Limit      int
	CategoryID *uuid.UUID
	Search     string
	SortBy     string
	Order      string
}

func NewProductQuery() ProductQuery {
	return ProductQuery{
		Page:  1,
		Limit: 10,
		Order: "asc",
	}
}

type ProductStatistics struct {
	TotalProducts int64 `json:"total_products"`
	ActiveSKU     int64 `json:"active_sku"`
}

type ProductWithRelations struct {
	Product models.Product
}
