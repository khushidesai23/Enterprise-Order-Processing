package inventory

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	auth gin.HandlerFunc,
) {
	inventory := router.Group("/inventory")
	{
		// CRUD
		inventory.POST("", auth, handler.CreateInventory)

		inventory.GET("", auth, handler.GetInventories)

		inventory.GET("/:productId", auth, handler.GetInventory)

		inventory.PUT("/:productId", auth, handler.UpdateInventory)

		// Stock Operations
		inventory.PATCH("/:productId/add-stock", auth, handler.AddStock)

		inventory.PATCH("/:productId/remove-stock", auth, handler.RemoveStock)

		inventory.PATCH("/:productId/reserve", auth, handler.ReserveStock)

		inventory.PATCH("/:productId/release", auth, handler.ReleaseReservedStock)

		inventory.PATCH("/:productId/confirm", auth, handler.ConfirmReservedStock)
	}
}
