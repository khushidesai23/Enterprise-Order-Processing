package product

import (
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

func ToProductModel(req CreateProductRequest) *models.Product {
	return &models.Product{
		Name:        req.Name,
		Description: req.Description,
		SKU:         req.SKU,
		Price:       req.Price,
		CategoryID:  req.CategoryID,
		Brand:       req.Brand,
		MRP:         req.MRP,
		ImageURL:    req.ImageURL,
		Images:      models.JSONList[string](req.Images),
		Specs:       models.JSONList[models.ProductSpec](req.Specs),
	}
}

func UpdateProductModel(product *models.Product, req UpdateProductRequest) {
	product.Name = req.Name
	product.Description = req.Description
	product.SKU = req.SKU
	product.Price = req.Price
	product.CategoryID = req.CategoryID

	if req.Brand != nil {
		product.Brand = *req.Brand
	}
	if req.MRP != nil {
		product.MRP = *req.MRP
	}
	if req.ImageURL != nil {
		product.ImageURL = *req.ImageURL
	}
	if req.Images != nil {
		product.Images = models.JSONList[string](*req.Images)
	}
	if req.Specs != nil {
		product.Specs = models.JSONList[models.ProductSpec](*req.Specs)
	}
}

func ToProductResponse(product *models.Product) ProductResponse {

	res := ProductResponse{
		ID:            product.ID,
		Name:          product.Name,
		Description:   product.Description,
		SKU:           product.SKU,
		Price:         product.Price,
		Brand:         product.Brand,
		MRP:           product.MRP,
		ImageURL:      product.ImageURL,
		Images:        nonNil(product.Images),
		Specs:         nonNil(product.Specs),
		RatingAverage: product.RatingAverage,
		RatingCount:   product.RatingCount,
	}

	// Category
	res.Category = CategoryResponse{
		ID:   product.Category.ID,
		Name: product.Category.Name,
	}

	// Inventory
	if product.Inventory != nil {
		res.Inventory = InventorySummary{
			AvailableQuantity: product.Inventory.AvailableQuantity,
			ReservedQuantity:  product.Inventory.ReservedQuantity,
		}
	}

	return res
}

func ToProductListResponse(product *models.Product) ProductListResponse {

	response := ProductListResponse{
		ID:            product.ID,
		Name:          product.Name,
		SKU:           product.SKU,
		Price:         product.Price,
		Brand:         product.Brand,
		MRP:           product.MRP,
		ImageURL:      product.ImageURL,
		RatingAverage: product.RatingAverage,
		RatingCount:   product.RatingCount,
		CategoryID:    product.CategoryID,
		Category:      product.Category.Name,
	}

	if product.Inventory != nil {
		response.Available = product.Inventory.AvailableQuantity
		response.Reserved = product.Inventory.ReservedQuantity
	}

	return response
}

func ToProductList(products []models.Product) []ProductListResponse {

	result := make([]ProductListResponse, 0, len(products))

	for i := range products {
		result = append(result, ToProductListResponse(&products[i]))
	}

	return result
}

// nonNil keeps JSON arrays as [] rather than null.
func nonNil[T any](items []T) []T {
	if items == nil {
		return []T{}
	}
	return items
}
