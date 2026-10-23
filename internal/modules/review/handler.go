package review

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/response"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/authz"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GET /products/:id/reviews
// @Summary List Product Reviews
// @Description One page of a product's reviews with the rating summary of all its reviews
// @Tags Reviews
// @Produce json
// @Param id path string true "Product ID"
// @Param rating query int false "Only reviews with this star rating (1-5)"
// @Param sort query string false "Sort order" Enums(recent, highest, lowest)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Reviews per page" default(10) maximum(100)
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /products/{id}/reviews [get]
func (h *Handler) ListProductReviews(c *gin.Context) {
	productID, ok := parseID(c, "id", "invalid product id")
	if !ok {
		return
	}

	page, limit, ok := parsePage(c)
	if !ok {
		return
	}

	rating, ok := parseRating(c)
	if !ok {
		return
	}

	result, err := h.service.ListProductReviews(c.Request.Context(), productID, rating, c.Query("sort"), page, limit)
	if err != nil {
		h.fail(c, err)
		return
	}

	response.OK(c, "reviews retrieved successfully", result)
}

// GET /products/:id/reviews/me
// @Summary Get My Review
// @Description The caller's review of a product
// @Tags Reviews
// @Security BearerAuth
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /products/{id}/reviews/me [get]
func (h *Handler) MyReview(c *gin.Context) {
	productID, ok := parseID(c, "id", "invalid product id")
	if !ok {
		return
	}

	principal, ok := authz.RequirePrincipal(c)
	if !ok {
		return
	}

	result, err := h.service.MyReview(c.Request.Context(), principal, productID)
	if err != nil {
		h.fail(c, err)
		return
	}

	response.OK(c, "review retrieved successfully", result)
}

// PUT /products/:id/reviews
// @Summary Write Review
// @Description Create or replace the caller's review of a product. Marked as a verified purchase when the caller has a paid order containing it.
// @Tags Reviews
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Param request body UpsertReviewRequest true "Review"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /products/{id}/reviews [put]
func (h *Handler) UpsertReview(c *gin.Context) {
	productID, ok := parseID(c, "id", "invalid product id")
	if !ok {
		return
	}

	var req UpsertReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	principal, ok := authz.RequirePrincipal(c)
	if !ok {
		return
	}

	result, err := h.service.UpsertReview(c.Request.Context(), principal, productID, req)
	if err != nil {
		h.fail(c, err)
		return
	}

	response.OK(c, "review saved successfully", result)
}

// DELETE /reviews/:id
// @Summary Delete Review
// @Description Delete a review (own reviews only, unless admin)
// @Tags Reviews
// @Security BearerAuth
// @Produce json
// @Param id path string true "Review ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /reviews/{id} [delete]
func (h *Handler) DeleteReview(c *gin.Context) {
	id, ok := parseID(c, "id", "invalid review id")
	if !ok {
		return
	}

	principal, ok := authz.RequirePrincipal(c)
	if !ok {
		return
	}

	if err := h.service.DeleteReview(c.Request.Context(), principal, id); err != nil {
		h.fail(c, err)
		return
	}

	response.OK(c, "review deleted successfully", nil)
}

// GET /reviews
// @Summary List Reviews
// @Description All reviews, newest first, for moderation (admin only)
// @Tags Reviews
// @Security BearerAuth
// @Produce json
// @Param rating query int false "Only reviews with this star rating (1-5)"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Reviews per page" default(10) maximum(100)
// @Success 200 {object} response.APIResponse
// @Failure 403 {object} response.APIResponse
// @Router /reviews [get]
func (h *Handler) ListReviews(c *gin.Context) {
	page, limit, ok := parsePage(c)
	if !ok {
		return
	}

	rating, ok := parseRating(c)
	if !ok {
		return
	}

	result, err := h.service.ListReviews(c.Request.Context(), rating, page, limit)
	if err != nil {
		h.fail(c, err)
		return
	}

	response.OK(c, "reviews retrieved successfully", result)
}

func (h *Handler) fail(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrProductNotFound), errors.Is(err, ErrReviewNotFound):
		response.Error(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrInvalidSort):
		response.Error(c, http.StatusBadRequest, err.Error())
	default:
		response.Error(c, http.StatusInternalServerError, err.Error())
	}
}

func parseID(c *gin.Context, param, message string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		response.Error(c, http.StatusBadRequest, message)
		return uuid.Nil, false
	}
	return id, true
}

func parsePage(c *gin.Context) (int, int, bool) {
	page, limit := 1, DefaultPageSize

	if value := c.Query("page"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100000 {
			response.Error(c, http.StatusBadRequest, "page must be a positive integer")
			return 0, 0, false
		}
		page = parsed
	}

	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > MaxPageSize {
			response.Error(c, http.StatusBadRequest, "limit must be an integer between 1 and 100")
			return 0, 0, false
		}
		limit = parsed
	}

	return page, limit, true
}

func parseRating(c *gin.Context) (int, bool) {
	value := c.Query("rating")
	if value == "" {
		return 0, true
	}

	rating, err := strconv.Atoi(value)
	if err != nil || rating < 1 || rating > 5 {
		response.Error(c, http.StatusBadRequest, "rating must be an integer between 1 and 5")
		return 0, false
	}

	return rating, true
}
