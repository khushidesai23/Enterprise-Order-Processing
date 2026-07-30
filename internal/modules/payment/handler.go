package payment

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
func (h *Handler) CreatePayment(c *gin.Context) {

	var req CreatePaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": err.Error(),
		})

		return
	}

	response, err := h.service.CreatePayment(
		c.Request.Context(),
		req,
	)
	if err != nil {

		switch {

		case errors.Is(err, ErrOrderNotFound):

			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})

			return

		case errors.Is(err, ErrOrderAlreadyPaid),
			errors.Is(err, ErrOrderCancelled),
			errors.Is(err, ErrPaymentAlreadyCompleted):

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
		"data":    response,
	})
}

// GET /payments/:id
func (h *Handler) GetPayment(c *gin.Context) {

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid payment id",
		})

		return
	}

	response, err := h.service.GetPayment(id)
	if err != nil {

		switch {

		case errors.Is(err, ErrPaymentNotFound):

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
		"data":    response,
	})
}

// GET /payments
func (h *Handler) GetPayments(c *gin.Context) {

	response, err := h.service.GetPayments()
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GET /payments/order/:orderId
func (h *Handler) GetPaymentByOrder(c *gin.Context) {

	orderID, err := uuid.Parse(c.Param("orderId"))
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid order id",
		})

		return
	}

	response, err := h.service.GetPaymentByOrder(orderID)
	if err != nil {

		switch {

		case errors.Is(err, ErrOrderNotFound),
			errors.Is(err, ErrPaymentNotFound):

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
		"data":    response,
	})
}

// GET /payments/summary
func (h *Handler) GetPaymentSummary(c *gin.Context) {

	response, err := h.service.GetPaymentSummary()
	if err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": err.Error(),
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// POST /payments/webhook
func (h *Handler) ProcessWebhook(c *gin.Context) {

	body, err := c.GetRawData()
	if err != nil {

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "failed to read request body",
		})

		return
	}

	signature := c.GetHeader("X-Razorpay-Signature")
	if signature == "" {

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "missing razorpay signature",
		})

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

			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"message": err.Error(),
			})

			return

		case errors.Is(err, ErrPaymentNotFound),
			errors.Is(err, ErrUnknownWebhookEvent):

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

	// IMPORTANT:
	// Razorpay expects HTTP 200 for successfully
	// processed webhooks. Returning 200 prevents
	// unnecessary retries.

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "webhook processed successfully",
	})
}

// POST /payments/:id/refund
// @Summary Refund Payment
// @Description Refund a payment by ID
// @Tags Payments
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

		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "invalid payment id",
		})

		return
	}

	response, err := h.service.RefundPayment(id)
	if err != nil {

		switch {

		case errors.Is(err, ErrPaymentNotFound):

			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"message": err.Error(),
			})

			return

		case errors.Is(err, ErrInvalidPaymentStatus),
			errors.Is(err, ErrPaymentNotPending):

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

	response.OK(c, "payment refunded successfully", refundResp)
}
