package product

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
// @Security BearerAuth
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

	product, err := h.service.CreateProduct(c.Request.Context(), req)
	if err != nil {

		switch {
		case errors.Is(err, ErrCategoryNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
			return

		case errors.Is(err, ErrProductSKUExists):
			response.Error(c, http.StatusConflict, err.Error())
			return

		case errors.Is(err, ErrInvalidMRP):
			response.Error(c, http.StatusBadRequest, err.Error())
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

	products, err := h.service.GetProducts(c.Request.Context())
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

	product, err := h.service.GetProduct(c.Request.Context(), id)
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

	products, err := h.service.GetProductsByCategory(c.Request.Context(), categoryID)
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
// @Security BearerAuth
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

	product, err := h.service.UpdateProduct(c.Request.Context(), id, req)
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

		case errors.Is(err, ErrInvalidMRP):
			response.Error(c, http.StatusBadRequest, err.Error())
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
// @Security BearerAuth
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

	err = h.service.DeleteProduct(c.Request.Context(), id)
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

// GET /products/search
// @Summary Search Products
// @Description Search the catalogue with filters, sorting, pagination and facets
// @Tags Products
// @Produce json
// @Param q query string false "Search words (all must match name, brand, category or SKU)"
// @Param category_id query string false "Category ID"
// @Param brand query []string false "Brands (repeatable)" collectionFormat(multi)
// @Param min_price query number false "Minimum price"
// @Param max_price query number false "Maximum price"
// @Param min_rating query number false "Minimum average rating (1-5)"
// @Param in_stock query bool false "Only products with available stock"
// @Param deals query bool false "Only discounted products (mrp above price)"
// @Param sort query string false "Sort order" Enums(relevance, featured, price_asc, price_desc, rating, reviews, newest, discount, bestselling)
// @Param page query int false "Page number" default(1) minimum(1)
// @Param limit query int false "Products per page" default(24) minimum(1) maximum(100)
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /products/search [get]
func (h *Handler) SearchProducts(c *gin.Context) {

	req, err := parseSearchRequest(c)
	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.SearchProducts(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidSort) {
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, "products retrieved successfully", result)
}

func parseSearchRequest(c *gin.Context) (SearchProductsRequest, error) {
	req := SearchProductsRequest{
		Query:  strings.TrimSpace(c.Query("q")),
		Brands: c.QueryArray("brand"),
		Sort:   c.Query("sort"),
		Page:   1,
		Limit:  DefaultSearchPageSize,
	}

	if value := c.Query("category_id"); value != "" {
		id, err := uuid.Parse(value)
		if err != nil {
			return req, ErrInvalidCategoryID
		}
		req.CategoryID = &id
	}

	for _, param := range []struct {
		name   string
		target **float64
	}{
		{"min_price", &req.MinPrice},
		{"max_price", &req.MaxPrice},
		{"min_rating", &req.MinRating},
	} {
		value := c.Query(param.name)
		if value == "" {
			continue
		}
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil || parsed < 0 {
			return req, fmt.Errorf("%s must be a non-negative number", param.name)
		}
		*param.target = &parsed
	}

	req.InStock = isTrue(c.Query("in_stock"))
	req.DealsOnly = isTrue(c.Query("deals"))

	if value := c.Query("page"); value != "" {
		page, err := strconv.Atoi(value)
		if err != nil || page < 1 || page > 10000 {
			return req, errors.New("page must be an integer between 1 and 10000")
		}
		req.Page = page
	}

	if value := c.Query("limit"); value != "" {
		limit, err := strconv.Atoi(value)
		if err != nil || limit < 1 || limit > MaxSearchPageSize {
			return req, fmt.Errorf("limit must be an integer between 1 and %d", MaxSearchPageSize)
		}
		req.Limit = limit
	}

	if req.Sort != "" && !searchSorts[req.Sort] {
		return req, ErrInvalidSort
	}

	return req, nil
}

func isTrue(value string) bool {
	return value == "true" || value == "1"
}
