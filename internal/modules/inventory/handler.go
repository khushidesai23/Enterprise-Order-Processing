package inventory

import (
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

// POST /inventory
func (h *Handler) CreateInventory(c *gin.Context) {
	var req CreateInventoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	inventory, err := h.service.CreateInventory(req)
	if err != nil {

		switch {
		case errors.Is(err, ErrProductNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return

		case errors.Is(err, ErrInventoryAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    inventory,
	})
}

// GET /inventory
func (h *Handler) GetInventories(c *gin.Context) {

	inventories, err := h.service.GetInventories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    inventories,
	})
}

// GET /inventory/:productId
func (h *Handler) GetInventory(c *gin.Context) {

	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": ErrInvalidInventoryID.Error(),
		})
		return
	}

	inventory, err := h.service.GetInventory(productID)
	if err != nil {

		switch {
		case errors.Is(err, ErrInventoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    inventory,
	})
}

// PUT /inventory/:productId
func (h *Handler) UpdateInventory(c *gin.Context) {

	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": ErrInvalidInventoryID.Error(),
		})
		return
	}

	var req UpdateInventoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	inventory, err := h.service.UpdateInventory(productID, req)
	if err != nil {

		switch {

		case errors.Is(err, ErrInventoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return

		case errors.Is(err, ErrNegativeStock):
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    inventory,
	})
}

// PATCH /inventory/:productId/add-stock
func (h *Handler) AddStock(c *gin.Context) {

	h.handleStockOperation(c, h.service.AddStock)
}

// PATCH /inventory/:productId/remove-stock
func (h *Handler) RemoveStock(c *gin.Context) {

	h.handleStockOperation(c, h.service.RemoveStock)
}

// PATCH /inventory/:productId/reserve
func (h *Handler) ReserveStock(c *gin.Context) {

	h.handleStockOperation(c, h.service.ReserveStock)
}

// PATCH /inventory/:productId/release
func (h *Handler) ReleaseReservedStock(c *gin.Context) {

	h.handleStockOperation(c, h.service.ReleaseReservedStock)
}

// PATCH /inventory/:productId/confirm
func (h *Handler) ConfirmReservedStock(c *gin.Context) {

	h.handleStockOperation(c, h.service.ConfirmReservedStock)
}

func (h *Handler) handleStockOperation(
	c *gin.Context,
	operation func(uuid.UUID, int) (*InventoryResponse, error),
) {

	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": ErrInvalidInventoryID.Error(),
		})
		return
	}

	var req StockOperationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	inventory, err := operation(productID, req.Quantity)
	if err != nil {

		switch {

		case errors.Is(err, ErrInventoryNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})

		case errors.Is(err, ErrInvalidQuantity),
			errors.Is(err, ErrInsufficientStock),
			errors.Is(err, ErrInsufficientReserved),
			errors.Is(err, ErrNegativeStock):
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": err.Error(),
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": err.Error(),
			})
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    inventory,
	})
}