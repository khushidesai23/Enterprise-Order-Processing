package inventory

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

// POST /inventory
// @Summary Create Inventory
// @Description Create a new inventory record
// @Tags Inventory
// @Accept json
// @Produce json
// @Param request body CreateInventoryRequest true "Inventory"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Router /inventory [post]
func (h *Handler) CreateInventory(c *gin.Context) {
	var req CreateInventoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	inventory, err := h.service.CreateInventory(req)
	if err != nil {

		switch {
		case errors.Is(err, ErrProductNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
			return

		case errors.Is(err, ErrInventoryAlreadyExists):
			response.Error(c, http.StatusConflict, err.Error())
			return

		case errors.Is(err, ErrInventoryAlreadyExists):
			response.Error(c, http.StatusConflict, err.Error())
			return
		}

		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Created(c, "inventory created successfully", inventory)
}

// GET /inventory
// @Summary Get Inventories
// @Description Get all inventory records
// @Tags Inventory
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /inventory [get]
func (h *Handler) GetInventories(c *gin.Context) {

	inventories, err := h.service.GetInventories()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "inventories retrieved successfully", inventories)
}

// GET /inventory/:productId
// @Summary Get Inventory by Product ID
// @Description Get inventory details by product ID
// @Tags Inventory
// @Accept json
// @Produce json
// @Param productId path string true "Product ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /inventory/{productId} [get]
func (h *Handler) GetInventory(c *gin.Context) {

	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, ErrInvalidInventoryID.Error())
		return
	}

	inventory, err := h.service.GetInventory(productID)
	if err != nil {

		switch {
		case errors.Is(err, ErrInventoryNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
			return
		}

		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, "inventory retrieved successfully", inventory)
}

// PUT /inventory/:productId
// @Summary Update Inventory by Product ID
// @Description Update inventory details by product ID
// @Tags Inventory
// @Accept json
// @Produce json
// @Param productId path string true "Product ID"
// @Param request body UpdateInventoryRequest true "Inventory Update"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /inventory/{productId} [put]
func (h *Handler) UpdateInventory(c *gin.Context) {

	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, ErrInvalidInventoryID.Error())
		return
	}

	var req UpdateInventoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	inventory, err := h.service.UpdateInventory(productID, req)
	if err != nil {

		switch {

		case errors.Is(err, ErrInventoryNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
			return

		case errors.Is(err, ErrNegativeStock):
			response.Error(c, http.StatusBadRequest, err.Error())
			return
		}

		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, "inventory updated successfully", inventory)
}

// PATCH /inventory/:productId/add-stock
// @Summary Add Stock to Inventory
// @Description Add stock to inventory by product ID
// @Tags Inventory
// @Accept json
// @Produce json
// @Param productId path string true "Product ID"
// @Param request body StockOperationRequest true "Stock Operation"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /inventory/{productId}/add-stock [patch]
func (h *Handler) AddStock(c *gin.Context) {

	h.handleStockOperation(c, h.service.AddStock)
}

// PATCH /inventory/:productId/remove-stock
// @Summary Remove Stock from Inventory
// @Description Remove stock from inventory by product ID
// @Tags Inventory
// @Accept json
// @Produce json
// @Param productId path string true "Product ID"
// @Param request body StockOperationRequest true "Stock Operation"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /inventory/{productId}/remove-stock [patch]
func (h *Handler) RemoveStock(c *gin.Context) {

	h.handleStockOperation(c, h.service.RemoveStock)
}

// PATCH /inventory/:productId/reserve
// @Summary Reserve Stock in Inventory
// @Description Reserve stock in inventory by product ID
// @Tags Inventory
// @Accept json
// @Produce json
// @Param productId path string true "Product ID"
// @Param request body StockOperationRequest true "Stock Operation"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /inventory/{productId}/reserve [patch]
func (h *Handler) ReserveStock(c *gin.Context) {

	h.handleStockOperation(c, h.service.ReserveStock)
}

// PATCH /inventory/:productId/release
// @Summary Release Reserved Stock in Inventory
// @Description Release reserved stock in inventory by product ID
// @Tags Inventory
// @Accept json
// @Produce json
// @Param productId path string true "Product ID"
// @Param request body StockOperationRequest true "Stock Operation"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /inventory/{productId}/release [patch]
func (h *Handler) ReleaseReservedStock(c *gin.Context) {

	h.handleStockOperation(c, h.service.ReleaseReservedStock)
}

// PATCH /inventory/:productId/confirm
// @Summary Confirm Reserved Stock in Inventory
// @Description Confirm reserved stock in inventory by product ID
// @Tags Inventory
// @Accept json
// @Produce json
// @Param productId path string true "Product ID"
// @Param request body StockOperationRequest true "Stock Operation"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /inventory/{productId}/confirm [patch]
func (h *Handler) ConfirmReservedStock(c *gin.Context) {

	h.handleStockOperation(c, h.service.ConfirmReservedStock)
}

// handleStockOperation is a helper method to handle stock operations like add, remove, reserve, release, and confirm.
func (h *Handler) handleStockOperation(
	c *gin.Context,
	operation func(uuid.UUID, int) (*InventoryResponse, error),
) {
	productID, err := uuid.Parse(c.Param("productId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, ErrInvalidInventoryID.Error())
		return
	}

	var req StockOperationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	inventory, err := operation(productID, req.Quantity)
	if err != nil {

		switch {

		case errors.Is(err, ErrInventoryNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
			return

		case errors.Is(err, ErrInvalidQuantity),
			errors.Is(err, ErrInsufficientStock),
			errors.Is(err, ErrInsufficientReserved),
			errors.Is(err, ErrNegativeStock):
			response.Error(c, http.StatusBadRequest, err.Error())
			return

		default:
			response.Error(c, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.OK(c, "stock operation completed successfully", inventory)
}