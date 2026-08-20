package product

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	auth gin.HandlerFunc,
) {
	products := router.Group("/products")
	{
		products.POST("", auth, handler.CreateProduct)

		products.GET("", handler.GetProducts)

		products.GET("/:id", handler.GetProduct)

		products.GET("/category/:categoryId", handler.GetProductsByCategory)

		products.PUT("/:id", auth, handler.UpdateProduct)

		products.DELETE("/:id", auth, handler.DeleteProduct)
	}
}
