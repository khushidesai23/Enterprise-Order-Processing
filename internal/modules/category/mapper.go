package category

import (
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

func ToCategoryModel(req CreateCategoryRequest) *models.Category {
	return &models.Category{
		Name:        req.Name,
		Description: req.Description,
	}
}

func UpdateCategoryModel(category *models.Category, req UpdateCategoryRequest) {
	category.Name = req.Name
	category.Description = req.Description
}

func ToCategoryResponse(category *models.Category) CategoryResponse {
	return CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Description: category.Description,
	}
}

func ToCategoryListResponse(category *models.Category) CategoryListResponse {
	var productCount int64

	if category.Products != nil {
		productCount = int64(len(category.Products))
	}

	return CategoryListResponse{
		ID:           category.ID,
		Name:         category.Name,
		Description:  category.Description,
		ProductCount: productCount,
	}
}

func ToCategoryList(categories []models.Category) []CategoryListResponse {
	response := make([]CategoryListResponse, 0, len(categories))

	for i := range categories {
		response = append(response, ToCategoryListResponse(&categories[i]))
	}

	return response
}
