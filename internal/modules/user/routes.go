package user

import "github.com/gin-gonic/gin"

// RegisterRoutes wires the user endpoints. Registration is public; users
// may read and update their own profile; listing, deleting and role
// changes require the admin role.
func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	auth gin.HandlerFunc,
	admin gin.HandlerFunc,
) {

	users := router.Group("/users")

	{
		users.POST("", handler.Create)

		users.GET("", auth, admin, handler.List)

		users.GET("/:id", auth, handler.GetByID)

		users.PUT("/:id", auth, handler.Update)

		users.PATCH("/:id/role", auth, admin, handler.UpdateRole)

		users.DELETE("/:id", auth, admin, handler.Delete)
	}
}
