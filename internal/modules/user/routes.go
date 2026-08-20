package user

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	auth gin.HandlerFunc,
) {

	users := router.Group("/users")

	{
		users.POST("", handler.Create)

		users.GET("", auth, handler.List)

		users.GET("/:id", auth, handler.GetByID)

		users.PUT("/:id", auth, handler.Update)

		users.DELETE("/:id", auth, handler.Delete)
	}
}
