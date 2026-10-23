package product

import "github.com/gin-gonic/gin"

// RegisterRoutes wires the product endpoints. Reads are public; writes
// require the admin role.
func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	auth gin.HandlerFunc,
	admin gin.HandlerFunc,
) {
	products := router.Group("/products")
	{
		products.POST("", auth, admin, handler.CreateProduct)

		products.GET("", handler.GetProducts)

		products.GET("/search", handler.SearchProducts)

		products.GET("/:id", handler.GetProduct)

		products.GET("/category/:categoryId", handler.GetProductsByCategory)

		products.PUT("/:id", auth, admin, handler.UpdateProduct)

		products.DELETE("/:id", auth, admin, handler.DeleteProduct)
	}
}
