package order

import (
	"context"
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

// POST /orders
// @Summary Create Order
// @Description Create a new order
// @Tags Orders
// @Accept json
// @Produce json
// @Param request body CreateOrderRequest true "Order"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {

	var req CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	order, err := h.service.CreateOrder(
		context.Background(),
		req,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound),
			errors.Is(err, ErrProductNotFound),
			errors.Is(err, ErrInventoryNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
			return

		case errors.Is(err, ErrInsufficientStock),
			errors.Is(err, ErrDuplicateProduct),
			errors.Is(err, ErrOrderItemsRequired):
			response.Error(c, http.StatusBadRequest, err.Error())
			return

		default:
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	response.Created(c, "order created successfully", order)
}

// GET /orders/:id
// @Summary Get Order by ID
// @Description Get order details by ID
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /orders/{id} [get]
func (h *Handler) GetOrder(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, ErrInvalidOrderID.Error())
		return
	}

	order, err := h.service.GetOrder(id)
	if err != nil {

		switch {

		case errors.Is(err, ErrOrderNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
			return

		default:
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	response.OK(c, "order retrieved successfully", order)
}

// GET /orders
// @Summary Get Orders
// @Description Get all orders
// @Tags Orders
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /orders [get]
func (h *Handler) GetOrders(c *gin.Context) {

	orders, err := h.service.GetOrders()
	if err != nil {

		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, "orders retrieved successfully", orders)
}

// GET /orders/user/:userId
// @Summary Get Orders by User ID
// @Description Get all orders for a specific user by user ID
// @Tags Orders
// @Accept json
// @Produce json
// @Param userId path string true "User ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /orders/user/{userId} [get]
func (h *Handler) GetOrdersByUser(c *gin.Context) {

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	orders, err := h.service.GetOrdersByUser(
		context.Background(),
		userID,
	)

	if err != nil {

		switch {

		case errors.Is(err, ErrUserNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
			return

		default:
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	response.OK(c, "orders retrieved successfully", orders)
}

// PATCH /orders/:id/status
// @Summary Update Order Status
// @Description Update the status of an order by ID
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Param status body UpdateOrderStatusRequest true "Order Status"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /orders/{id}/status [patch]
func (h *Handler) UpdateOrderStatus(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, ErrInvalidOrderID.Error())
		return
	}

	var req UpdateOrderStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	order, err := h.service.UpdateOrderStatus(id, req)
	if err != nil {

		switch {

		case errors.Is(err, ErrOrderNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
			return

		case errors.Is(err, ErrInvalidOrderStatus),
			errors.Is(err, ErrOrderAlreadyCancelled),
			errors.Is(err, ErrOrderAlreadyCompleted):
			response.Error(c, http.StatusBadRequest, err.Error())
			return

		default:
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	response.OK(c, "order status updated successfully", order)
}

// PATCH /orders/:id/cancel
// @Summary Cancel Order
// @Description Cancel an order by ID
// @Tags Orders
// @Accept json
// @Produce json
// @Param id path string true "Order ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /orders/{id}/cancel [patch]
func (h *Handler) CancelOrder(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, ErrInvalidOrderID.Error())
		return
	}

	order, err := h.service.CancelOrder(id)
	if err != nil {

		switch {

		case errors.Is(err, ErrOrderNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
			return

		case errors.Is(err, ErrOrderAlreadyCancelled),
			errors.Is(err, ErrOrderAlreadyCompleted),
			errors.Is(err, ErrOrderCannotBeCancelled),
			errors.Is(err, ErrInventoryNotFound),
			errors.Is(err, ErrInsufficientReserved):
			response.Error(c, http.StatusBadRequest, err.Error())
			return

		default:
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	response.OK(c, "Order cancelled successfully", order)
}