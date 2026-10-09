package middleware

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestTimeout bounds the request context. Database calls made with the
// request context are cancelled at the deadline, so a stalled dependency
// produces a 504 the client receives (see response.Error) instead of a
// handler that outlives http.Server.WriteTimeout and has its connection
// dropped after logging a "successful" status.
//
// timeout must be shorter than the server's WriteTimeout.
func RequestTimeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
