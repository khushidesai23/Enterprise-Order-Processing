package inventory

import (
	"github.com/google/uuid"
)

type CreateInventoryRequest struct {
	ProductID         uuid.UUID `json:"product_id" binding:"required"`
	AvailableQuantity int       `json:"available_quantity" binding:"gte=0"`
}

type UpdateInventoryRequest struct {
	AvailableQuantity int `json:"available_quantity" binding:"gte=0"`
	ReservedQuantity  int `json:"reserved_quantity" binding:"gte=0"`
}

type StockOperationRequest struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

type InventoryResponse struct {
	ProductID         uuid.UUID `json:"product_id"`
	AvailableQuantity int       `json:"available_quantity"`
	ReservedQuantity  int       `json:"reserved_quantity"`
}

type InventoryListResponse struct {
	ProductID         uuid.UUID `json:"product_id"`
	ProductName       string    `json:"product_name"`
	SKU               string    `json:"sku"`
	AvailableQuantity int       `json:"available_quantity"`
	ReservedQuantity  int       `json:"reserved_quantity"`
}

type InventorySummary struct {
	TotalProducts     int64 `json:"total_products"`
	AvailableProducts int64 `json:"available_products"`
	OutOfStock        int64 `json:"out_of_stock"`
}