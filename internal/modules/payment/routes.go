package payment

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
) {
	payments := router.Group("/payments")
	{
		// Create Payment
		payments.POST("", handler.CreatePayment)

		// List Payments
		payments.GET("", handler.GetPayments)

		// Payment Summary
		payments.GET("/summary", handler.GetPaymentSummary)

		// Get Payment
		payments.GET("/:id", handler.GetPayment)

		// Get Payment By Order
		payments.GET("/order/:orderId", handler.GetPaymentByOrder)

		// Razorpay Webhook
		payments.POST("/webhook", handler.ProcessWebhook)

		// Refund
		payments.POST("/:id/refund", handler.RefundPayment)
	}
}