package inventory

import "github.com/gin-gonic/gin"

// RegisterRoutes wires the inventory endpoints. Any authenticated user may
// read stock; every write requires the admin role.
func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	auth gin.HandlerFunc,
	admin gin.HandlerFunc,
) {
	inventory := router.Group("/inventory")
	{
		// CRUD
		inventory.POST("", auth, admin, handler.CreateInventory)

		inventory.GET("", auth, handler.GetInventories)

		inventory.GET("/:productId", auth, handler.GetInventory)

		inventory.PUT("/:productId", auth, admin, handler.UpdateInventory)

		// Stock Operations
		inventory.PATCH("/:productId/add-stock", auth, admin, handler.AddStock)

		inventory.PATCH("/:productId/remove-stock", auth, admin, handler.RemoveStock)

		inventory.PATCH("/:productId/reserve", auth, admin, handler.ReserveStock)

		inventory.PATCH("/:productId/release", auth, admin, handler.ReleaseReservedStock)

		inventory.PATCH("/:productId/confirm", auth, admin, handler.ConfirmReservedStock)
	}
}
