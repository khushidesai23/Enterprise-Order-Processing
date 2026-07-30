package product

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// POST /products
// @Summary Create Product
// @Description Create a new product
// @Tags Products
// @Accept json
// @Produce json
// @Param request body CreateProductRequest true "Product"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Router /products [post]
func (h *Handler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	product, err := h.service.CreateProduct(req)
	if err != nil {

		switch {
		case errors.Is(err, ErrCategoryNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
			return

		case errors.Is(err, ErrProductSKUExists):
			response.Error(c, http.StatusConflict, err.Error())
			return
		}

		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.Created(c, "product created successfully", product)
}

// GET /products
// @Summary Get Products
// @Description Get all products
// @Tags Products
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /products [get]
func (h *Handler) GetProducts(c *gin.Context) {

	products, err := h.service.GetProducts()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "products retrieved successfully", products)
}

// GET /products/:id
// @Summary Get Product by ID
// @Description Get product details by ID
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /products/{id} [get]
func (h *Handler) GetProduct(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, ErrInvalidProductID.Error())
		return
	}

	product, err := h.service.GetProduct(id)
	if err != nil {

		switch {
		case errors.Is(err, ErrProductNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}

		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, "product retrieved successfully", product)
}

// GET /products/category/:categoryId
// @Summary Get Products by Category
// @Description Get products by category ID
// @Tags Products
// @Accept json
// @Produce json
// @Param categoryId path string true "Category ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /products/category/{categoryId} [get]
func (h *Handler) GetProductsByCategory(c *gin.Context) {

	categoryID, err := uuid.Parse(c.Param("categoryId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, ErrInvalidCategoryID.Error())
		return
	}

	products, err := h.service.GetProductsByCategory(categoryID)
	if err != nil {

		switch {
		case errors.Is(err, ErrCategoryNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}

		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, "products retrieved successfully", products)
}

// PUT /products/:id
// @Summary Update Product
// @Description Update product details by ID
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param request body UpdateProductRequest true "Product"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /products/{id} [put]
func (h *Handler) UpdateProduct(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, ErrInvalidProductID.Error())
		return
	}

	var req UpdateProductRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	product, err := h.service.UpdateProduct(id, req)
	if err != nil {

		switch {

		case errors.Is(err, ErrProductNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
			return

		case errors.Is(err, ErrCategoryNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
			return

		case errors.Is(err, ErrProductSKUExists):
			response.Error(c, http.StatusConflict, err.Error())
			return
		}

		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, "product updated successfully", product)
}

// DELETE /products/:id
// @Summary Delete Product
// @Description Delete product by ID
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /products/{id} [delete]
func (h *Handler) DeleteProduct(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, ErrInvalidProductID.Error())
		return
	}

	err = h.service.DeleteProduct(id)
	if err != nil {

		switch {
		case errors.Is(err, ErrProductNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}

		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, "Product deleted successfully", nil)
}