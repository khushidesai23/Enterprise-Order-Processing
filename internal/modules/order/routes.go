package order

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	auth gin.HandlerFunc,
) {
	orders := router.Group("/orders")
	{
		// Create Order
		orders.POST("", auth, handler.CreateOrder)

		// List Orders
		orders.GET("", auth, handler.GetOrders)

		// Get Single Order
		orders.GET("/:id", auth, handler.GetOrder)

		// Get Orders By User
		orders.GET("/me", auth, handler.GetOrdersByUser)

		// Update Status
		orders.PATCH("/:id/status", auth, handler.UpdateOrderStatus)

		// Cancel Order
		orders.PATCH("/:id/cancel", auth, handler.CancelOrder)
	}
}