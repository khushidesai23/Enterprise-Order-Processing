package category

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	auth gin.HandlerFunc,
) {
	categories := router.Group("/categories")
	{
		categories.POST("", auth, handler.CreateCategory)

		categories.GET("", handler.GetCategories)

		categories.GET("/:id", handler.GetCategory)

		categories.PUT("/:id", auth, handler.UpdateCategory)

		categories.DELETE("/:id", auth, handler.DeleteCategory)
	}
}
