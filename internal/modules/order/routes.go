package order

import "github.com/gin-gonic/gin"

// RegisterRoutes wires the order endpoints. auth authenticates the caller;
// admin additionally requires the admin role. Ownership of individual
// orders is enforced by the service.
func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	auth gin.HandlerFunc,
	admin gin.HandlerFunc,
) {
	orders := router.Group("/orders")
	{
		// Create Order
		orders.POST("", auth, handler.CreateOrder)

		// List Orders (all for admins, own for customers)
		orders.GET("", auth, handler.GetOrders)

		// Get Orders By User
		orders.GET("/me", auth, handler.GetOrdersByUser)

		// Get Single Order
		orders.GET("/:id", auth, handler.GetOrder)

		// Update Status (fulfilment is an admin operation)
		orders.PATCH("/:id/status", auth, admin, handler.UpdateOrderStatus)

		// Cancel Order
		orders.PATCH("/:id/cancel", auth, handler.CancelOrder)
	}
}
