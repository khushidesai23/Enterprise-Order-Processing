package review

import "github.com/gin-gonic/gin"

// RegisterRoutes wires the review endpoints. Reading a product's reviews
// is public; writing needs an account; listing every review is for admins.
// Ownership of individual reviews is enforced by the service.
func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	auth gin.HandlerFunc,
	admin gin.HandlerFunc,
) {
	// The :id segment must keep the product routes' parameter name.
	productReviews := router.Group("/products/:id/reviews")
	{
		productReviews.GET("", handler.ListProductReviews)

		productReviews.GET("/me", auth, handler.MyReview)

		productReviews.PUT("", auth, handler.UpsertReview)
	}

	reviews := router.Group("/reviews")
	{
		reviews.GET("", auth, admin, handler.ListReviews)

		reviews.DELETE("/:id", auth, handler.DeleteReview)
	}
}
