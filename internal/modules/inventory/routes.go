package inventory

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
) {
	inventory := router.Group("/inventory")
	{
		// CRUD
		inventory.POST("", handler.CreateInventory)

		inventory.GET("", handler.GetInventories)

		inventory.GET("/:productId", handler.GetInventory)

		inventory.PUT("/:productId", handler.UpdateInventory)

		// Stock Operations
		inventory.PATCH("/:productId/add-stock", handler.AddStock)

		inventory.PATCH("/:productId/remove-stock", handler.RemoveStock)

		inventory.PATCH("/:productId/reserve", handler.ReserveStock)

		inventory.PATCH("/:productId/release", handler.ReleaseReservedStock)

		inventory.PATCH("/:productId/confirm", handler.ConfirmReservedStock)
	}
}