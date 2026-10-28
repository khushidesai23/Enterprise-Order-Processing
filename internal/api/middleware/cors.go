package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS allows the given browser origins to call the API. With none, every
// cross-origin request is refused (cors.New rejects an empty list).
func CORS(allowedOrigins []string) gin.HandlerFunc {

	var allowNone func(string) bool
	if len(allowedOrigins) == 0 {
		allowNone = func(string) bool { return false }
	}

	return cors.New(cors.Config{
		AllowOrigins:    allowedOrigins,
		AllowOriginFunc: allowNone,

		AllowMethods: []string{
			"GET",
			"POST",
			"PUT",
			"PATCH",
			"DELETE",
			"OPTIONS",
		},

		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
		},

		ExposeHeaders: []string{
			"Content-Length",
		},

		AllowCredentials: true,

		MaxAge: 12 * time.Hour,
	})
}
