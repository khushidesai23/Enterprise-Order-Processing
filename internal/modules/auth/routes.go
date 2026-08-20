package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
) {

	auth := router.Group("/auth")

	{
		auth.POST(
			"/login",
			handler.Login,
		)

		auth.GET(
			"/me",
			handler.Me,
		)
	}
}
