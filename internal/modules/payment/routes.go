package payment

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	auth gin.HandlerFunc,
) {
	payments := router.Group("/payments")
	{
		// Create Payment
		payments.POST("", auth, handler.CreatePayment)

		// List Payments
		payments.GET("", auth, handler.GetPayments)

		// Payment Summary
		payments.GET("/summary", auth, handler.GetPaymentSummary)

		// Get Payment
		payments.GET("/:id", auth, handler.GetPayment)

		// Get Payment By Order
		payments.GET("/order/:orderId", auth, handler.GetPaymentByOrder)

		// Razorpay Webhook
		payments.POST("/webhook", handler.ProcessWebhook)

		// Refund
		payments.POST("/:id/refund", auth, handler.RefundPayment)
	}
}