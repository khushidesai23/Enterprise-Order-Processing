package user

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
) {

	users := router.Group("/users")

	{
		users.POST("", handler.Create)

		users.GET("", handler.List)

		users.GET("/:id", handler.GetByID)

		users.PUT("/:id", handler.Update)

		users.DELETE("/:id", handler.Delete)
	}
}