package category

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
) {
	categories := router.Group("/categories")
	{
		categories.POST("", handler.CreateCategory)

		categories.GET("", handler.GetCategories)

		categories.GET("/:id", handler.GetCategory)

		categories.PUT("/:id", handler.UpdateCategory)

		categories.DELETE("/:id", handler.DeleteCategory)
	}
}