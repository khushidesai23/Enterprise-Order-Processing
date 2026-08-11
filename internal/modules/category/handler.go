package category

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

// POST /categories
// @Summary Create Category
// @Description Create a new category
// @Tags Categories
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateCategoryRequest true "Category"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Router /categories [post]
func (h *Handler) CreateCategory(c *gin.Context) {

	var req CreateCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	category, err := h.service.CreateCategory(
		c.Request.Context(),
		req,
	)
	if err != nil {

		switch {
		case errors.Is(err, ErrCategoryNameExists):
			response.Error(
				c,
				http.StatusConflict,
				err.Error(),
			)
			return
		}

		response.Error(
			c,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	response.Created(
		c,
		"category created successfully",
		category,
	)
}

// GET /categories
// @Summary Get Categories
// @Description Get all categories
// @Tags Categories
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /categories [get]
func (h *Handler) GetCategories(c *gin.Context) {

	categories, err := h.service.GetCategories(
		c.Request.Context(),
	)
	if err != nil {
		response.Error(
			c,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	response.OK(
		c,
		"categories retrieved successfully",
		categories,
	)
}

// GET /categories/:id
// @Summary Get Category by ID
// @Description Get category details by ID
// @Tags Categories
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /categories/{id} [get]
func (h *Handler) GetCategory(c *gin.Context) {

	id, err := uuid.Parse(
		c.Param("id"),
	)
	if err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			ErrInvalidCategoryID.Error(),
		)
		return
	}

	category, err := h.service.GetCategory(
		c.Request.Context(),
		id,
	)
	if err != nil {

		switch {
		case errors.Is(err, ErrCategoryNotFound):
			response.Error(
				c,
				http.StatusNotFound,
				err.Error(),
			)
			return
		}

		response.Error(
			c,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	response.OK(
		c,
		"category retrieved successfully",
		category,
	)
}

// PUT /categories/:id
// @Summary Update Category
// @Description Update category details by ID
// @Tags Categories
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param request body UpdateCategoryRequest true "Updated Category"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /categories/{id} [put]
func (h *Handler) UpdateCategory(c *gin.Context) {

	id, err := uuid.Parse(
		c.Param("id"),
	)
	if err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			ErrInvalidCategoryID.Error(),
		)
		return
	}

	var req UpdateCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			err.Error(),
		)
		return
	}

	category, err := h.service.UpdateCategory(
		c.Request.Context(),
		id,
		req,
	)
	if err != nil {

		switch {

		case errors.Is(err, ErrCategoryNotFound):
			response.Error(
				c,
				http.StatusNotFound,
				err.Error(),
			)
			return

		case errors.Is(err, ErrCategoryNameExists):
			response.Error(
				c,
				http.StatusConflict,
				err.Error(),
			)
			return
		}

		response.Error(
			c,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	response.OK(
		c,
		"category updated successfully",
		category,
	)
}

// DELETE /categories/:id
// @Summary Delete Category
// @Description Delete category by ID
// @Tags Categories
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /categories/{id} [delete]
func (h *Handler) DeleteCategory(c *gin.Context) {

	id, err := uuid.Parse(
		c.Param("id"),
	)
	if err != nil {
		response.Error(
			c,
			http.StatusBadRequest,
			ErrInvalidCategoryID.Error(),
		)
		return
	}

	err = h.service.DeleteCategory(
		c.Request.Context(),
		id,
	)
	if err != nil {

		switch {

		case errors.Is(err, ErrCategoryNotFound):
			response.Error(
				c,
				http.StatusNotFound,
				err.Error(),
			)
			return

		case errors.Is(err, ErrCategoryInUse):
			response.Error(
				c,
				http.StatusConflict,
				err.Error(),
			)
			return
		}

		response.Error(
			c,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	response.OK(
		c,
		"category deleted successfully",
		nil,
	)
}