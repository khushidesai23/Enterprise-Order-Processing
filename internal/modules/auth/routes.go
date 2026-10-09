package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	auth gin.HandlerFunc,
) {
	authGroup := router.Group("/auth")
	{
		// Login
		authGroup.POST("/login", handler.Login)

		// Current User
		authGroup.GET("/me", auth, handler.Me)
	}
}
