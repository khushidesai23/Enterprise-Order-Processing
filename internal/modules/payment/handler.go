package payment

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

func NewHandler(
	service *Service,
) *Handler {

	return &Handler{
		service: service,
	}
}

// POST /payments
// @Summary Create Payment
// @Description Create a new payment for an order
// @Tags Payments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreatePaymentRequest true "Payment"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /payments [post]
func (h *Handler) CreatePayment(c *gin.Context) {

	var req CreatePaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	paymentResp, err := h.service.CreatePayment(
		c.Request.Context(),
		req,
	)
	if err != nil {

		switch {
		case errors.Is(err, ErrOrderNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
			return

		case errors.Is(err, ErrOrderAlreadyPaid),
			errors.Is(err, ErrOrderCancelled),
			errors.Is(err, ErrPaymentAlreadyCompleted):

			response.Error(c, http.StatusBadRequest, err.Error())
			return

		default:
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	response.Created(c, "payment created successfully", paymentResp)
}

// GET /payments/:id
// @Summary Get Payment
// @Description Get a payment by ID
// @Tags Payments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /payments/{id} [get]
func (h *Handler) GetPayment(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid payment id")
		return
	}

	paymentResp, err := h.service.GetPayment(id)
	if err != nil {

		switch {
		case errors.Is(err, ErrPaymentNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
			return
		default:
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
	}
	response.OK(c, "payment retrieved successfully", paymentResp)
}

// GET /payments
// @Summary Get Payments
// @Description Get all payments
// @Tags Payments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /payments [get]
func (h *Handler) GetPayments(c *gin.Context) {

	paymentsResp, err := h.service.GetPayments()
	if err != nil {

		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	response.OK(c, "payments retrieved successfully", paymentsResp)
}

// GET /payments/order/:orderId
// @Summary Get Payment by Order ID
// @Description Get a payment by order ID
// @Tags Payments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param orderId path string true "Order ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /payments/order/{orderId} [get]
func (h *Handler) GetPaymentByOrder(c *gin.Context) {

	orderID, err := uuid.Parse(c.Param("orderId"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid order id")
		return
	}

	paymentResp, err := h.service.GetPaymentByOrder(orderID)
	if err != nil {
		switch {
		case errors.Is(err, ErrOrderNotFound),
			errors.Is(err, ErrPaymentNotFound):

			response.Error(c, http.StatusNotFound, err.Error())
			return

		default:
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
	}
	response.OK(c, "payment retrieved successfully", paymentResp)
}

// GET /payments/summary
// @Summary Get Payment Summary
// @Description Get a summary of payments
// @Tags Payments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /payments/summary [get]
func (h *Handler) GetPaymentSummary(c *gin.Context) {

	paymentSummaryResp, err := h.service.GetPaymentSummary()

	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.OK(c, "payment summary retrieved successfully", paymentSummaryResp)
}

// POST /payments/webhook
// @Summary Process Payment Webhook
// @Description Process a payment webhook from Razorpay
// @Tags Payments
// @Accept json
// @Produce json
// @Param request body ProcessWebhookRequest true "Webhook Request"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /payments/webhook [post]
func (h *Handler) ProcessWebhook(c *gin.Context) {

	body, err := c.GetRawData()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "failed to read request body")
		return
	}

	signature := c.GetHeader("X-Razorpay-Signature")
	if signature == "" {
		response.Error(c, http.StatusBadRequest, "missing razorpay signature")
		return
	}

	eventID := c.GetHeader("X-Razorpay-Event-Id")
	if eventID == "" {
		eventID = c.GetHeader("X-Razorpay-Request-Id")
	}

	req := ProcessWebhookRequest{
		Body:      body,
		Signature: signature,
		EventID:   eventID,
	}

	if err := h.service.ProcessWebhook(
		c.Request.Context(),
		req,
	); err != nil {

		switch {

		case errors.Is(err, ErrInvalidSignature):
			response.Error(c, http.StatusUnauthorized, err.Error())
			return

		case errors.Is(err, ErrPaymentNotFound),
			errors.Is(err, ErrUnknownWebhookEvent):

			response.Error(c, http.StatusBadRequest, err.Error())
			return

		default:

			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// IMPORTANT:
	// Razorpay expects HTTP 200 for successfully processed webhooks. Returning 200 prevents unnecessary retries.

	response.OK(c, "webhook processed successfully", nil)
}

// POST /payments/:id/refund
// @Summary Refund Payment
// @Description Refund a payment by ID
// @Tags Payments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 500 {object} response.APIResponse
// @Router /payments/{id}/refund [post]
func (h *Handler) RefundPayment(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid payment id")
		return
	}

	refundResp, err := h.service.RefundPayment(id)
	if err != nil {

		switch {
		case errors.Is(err, ErrPaymentNotFound):
			response.Error(c, http.StatusNotFound, err.Error())
			return

		case errors.Is(err, ErrInvalidPaymentStatus),
			errors.Is(err, ErrPaymentNotPending):

			response.Error(c, http.StatusBadRequest, err.Error())
			return

		default:
			response.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
	}

	response.OK(c, "payment refunded successfully", refundResp)
}
