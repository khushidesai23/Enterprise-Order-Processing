package order

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
) {
	orders := router.Group("/orders")
	{
		// Create Order
		orders.POST("", handler.CreateOrder)

		// List Orders
		orders.GET("", handler.GetOrders)

		// Get Single Order
		orders.GET("/:id", handler.GetOrder)

		// Get Orders By User
		orders.GET("/user/:userId", handler.GetOrdersByUser)

		// Update Status
		orders.PATCH("/:id/status", handler.UpdateOrderStatus)

		// Cancel Order
		orders.PATCH("/:id/cancel", handler.CancelOrder)
	}
}