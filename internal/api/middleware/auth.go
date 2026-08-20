package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/response"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/auth"
)

func AuthMiddleware(
	jwtManager *auth.JWTManager,
) gin.HandlerFunc {

	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {

			response.Error(
				c,
				http.StatusUnauthorized,
				auth.ErrMissingToken.Error(),
			)

			c.Abort()

			return
		}

		const bearerPrefix = "Bearer "

		if !strings.HasPrefix(authHeader, bearerPrefix) {

			response.Error(
				c,
				http.StatusUnauthorized,
				auth.ErrInvalidToken.Error(),
			)

			c.Abort()

			return
		}

		token := strings.TrimSpace(
			strings.TrimPrefix(authHeader, bearerPrefix),
		)

		if token == "" {

			response.Error(
				c,
				http.StatusUnauthorized,
				auth.ErrMissingToken.Error(),
			)

			c.Abort()

			return
		}

		claims, err := jwtManager.VerifyToken(token)

		if err != nil {

			response.Error(
				c,
				http.StatusUnauthorized,
				err.Error(),
			)

			c.Abort()

			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)

		c.Next()
	}
}
