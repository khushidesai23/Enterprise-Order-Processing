package category

import "github.com/gin-gonic/gin"

// RegisterRoutes wires the category endpoints. Reads are public; writes
// require the admin role.
func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	auth gin.HandlerFunc,
	admin gin.HandlerFunc,
) {
	categories := router.Group("/categories")
	{
		categories.POST("", auth, admin, handler.CreateCategory)

		categories.GET("", handler.GetCategories)

		categories.GET("/:id", handler.GetCategory)

		categories.PUT("/:id", auth, admin, handler.UpdateCategory)

		categories.DELETE("/:id", auth, admin, handler.DeleteCategory)
	}
}
