package payment

import "github.com/gin-gonic/gin"

// RegisterRoutes wires the payment endpoints. auth authenticates the
// caller; admin additionally requires the admin role. Ownership of
// individual payments is enforced by the service.
func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	auth gin.HandlerFunc,
	admin gin.HandlerFunc,
) {
	payments := router.Group("/payments")
	{
		// Create Payment
		payments.POST("", auth, handler.CreatePayment)

		// List Payments
		payments.GET("", auth, admin, handler.GetPayments)

		// Payment Summary
		payments.GET("/summary", auth, admin, handler.GetPaymentSummary)

		// Get Payment
		payments.GET("/:id", auth, handler.GetPayment)

		// Get Payment By Order
		payments.GET("/order/:orderId", auth, handler.GetPaymentByOrder)

		// Razorpay Webhook (authenticated by its signature)
		payments.POST("/webhook", handler.ProcessWebhook)

		// Refund
		payments.POST("/:id/refund", auth, admin, handler.RefundPayment)
	}
}
