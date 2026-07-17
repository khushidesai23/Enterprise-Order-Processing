package order

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
func (h *Handler) CreateOrder(c *gin.Context) {

	var req CreateOrderRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	order, err := h.service.CreateOrder(
		context.Background(),
		req,
	)

	if err != nil {

		switch {

		case errors.Is(err, ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return

		case errors.Is(err, ErrProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return

		case errors.Is(err, ErrInventoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return

		case errors.Is(err, ErrInsufficientStock):
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return

		case errors.Is(err, ErrDuplicateProduct):
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return

		case errors.Is(err, ErrOrderItemsRequired):
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    order,
	})
}

// GET /orders/:id
func (h *Handler) GetOrder(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": ErrInvalidOrderID.Error(),
		})
		return
	}

	order, err := h.service.GetOrder(id)
	if err != nil {

		switch {

		case errors.Is(err, ErrOrderNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    order,
	})
}

// GET /orders
func (h *Handler) GetOrders(c *gin.Context) {

	orders, err := h.service.GetOrders()
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    orders,
	})
}

// GET /orders/user/:userId
func (h *Handler) GetOrdersByUser(c *gin.Context) {

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid user id",
		})

		return
	}

	orders, err := h.service.GetOrdersByUser(
		context.Background(),
		userID,
	)

	if err != nil {

		switch {

		case errors.Is(err, ErrUserNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    orders,
	})
}

// PATCH /orders/:id/status
func (h *Handler) UpdateOrderStatus(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": ErrInvalidOrderID.Error(),
		})
		return
	}

	var req UpdateOrderStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	order, err := h.service.UpdateOrderStatus(id, req)
	if err != nil {

		switch {

		case errors.Is(err, ErrOrderNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return

		case errors.Is(err, ErrInvalidOrderStatus),
			errors.Is(err, ErrOrderAlreadyCancelled),
			errors.Is(err, ErrOrderAlreadyCompleted):
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    order,
	})
}

// PATCH /orders/:id/cancel
func (h *Handler) CancelOrder(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": ErrInvalidOrderID.Error(),
		})
		return
	}

	order, err := h.service.CancelOrder(id)
	if err != nil {

		switch {

		case errors.Is(err, ErrOrderNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return

		case errors.Is(err, ErrOrderAlreadyCancelled),
			errors.Is(err, ErrOrderAlreadyCompleted),
			errors.Is(err, ErrOrderCannotBeCancelled),
			errors.Is(err, ErrInventoryNotFound),
			errors.Is(err, ErrInsufficientReserved):
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Order cancelled successfully",
		"data":    order,
	})
}