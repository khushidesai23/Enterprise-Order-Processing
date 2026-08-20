package inventory

import (
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

func ToInventoryModel(req CreateInventoryRequest) *models.Inventory {
	return &models.Inventory{
		ProductID:         req.ProductID,
		AvailableQuantity: req.AvailableQuantity,
		ReservedQuantity:  0,
	}
}

func UpdateInventoryModel(
	inventory *models.Inventory,
	req UpdateInventoryRequest,
) {
	inventory.AvailableQuantity = req.AvailableQuantity
	inventory.ReservedQuantity = req.ReservedQuantity
}

func ToInventoryResponse(
	inventory *models.Inventory,
) InventoryResponse {

	return InventoryResponse{
		ProductID:         inventory.ProductID,
		AvailableQuantity: inventory.AvailableQuantity,
		ReservedQuantity:  inventory.ReservedQuantity,
	}
}

func ToInventoryListResponse(
	inventory *models.Inventory,
) InventoryListResponse {

	response := InventoryListResponse{
		ProductID:         inventory.ProductID,
		AvailableQuantity: inventory.AvailableQuantity,
		ReservedQuantity:  inventory.ReservedQuantity,
	}

	// Product information (requires Preload("Product"))
	if inventory.Product.ID != inventory.ProductID {
		response.ProductName = inventory.Product.Name
		response.SKU = inventory.Product.SKU
	}

	return response
}

func ToInventoryList(
	inventories []models.Inventory,
) []InventoryListResponse {

	response := make([]InventoryListResponse, 0, len(inventories))

	for i := range inventories {
		response = append(
			response,
			ToInventoryListResponse(&inventories[i]),
		)
	}

	return response
}

func ToInventorySummary(
	inventories []models.Inventory,
) InventorySummary {

	var summary InventorySummary

	summary.TotalProducts = int64(len(inventories))

	for _, inventory := range inventories {

		if inventory.AvailableQuantity > 0 {
			summary.AvailableProducts++
		} else {
			summary.OutOfStock++
		}
	}

	return summary
}
