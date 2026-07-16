package product

import "github.com/gin-gonic/gin"

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
) {
	products := router.Group("/products")
	{
		products.POST("", handler.CreateProduct)

		products.GET("", handler.GetProducts)

		products.GET("/:id", handler.GetProduct)

		products.GET("/category/:categoryId", handler.GetProductsByCategory)

		products.PUT("/:id", handler.UpdateProduct)

		products.DELETE("/:id", handler.DeleteProduct)
	}
}