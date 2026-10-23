package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/response"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/authz"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
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

		userID, err := uuid.Parse(claims.UserID)

		if err != nil {

			response.Error(
				c,
				http.StatusUnauthorized,
				auth.ErrInvalidToken.Error(),
			)

			c.Abort()

			return
		}

		// Tokens issued before roles existed, or carrying an unknown
		// role, get the least privileged role.
		role := models.Role(claims.Role)
		if !role.Valid() {
			role = models.RoleCustomer
		}

		authz.SetPrincipal(c, authz.Principal{
			UserID: userID,
			Role:   role,
		})

		c.Set("user_id", claims.UserID)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// RequireRole allows the request through only when the authenticated
// caller has one of roles. It must run after AuthMiddleware.
func RequireRole(
	roles ...models.Role,
) gin.HandlerFunc {

	return func(c *gin.Context) {

		principal, ok := authz.FromContext(c)

		if !ok {

			response.Error(
				c,
				http.StatusUnauthorized,
				auth.ErrMissingToken.Error(),
			)

			c.Abort()

			return
		}

		for _, role := range roles {

			if principal.Role == role {
				c.Next()
				return
			}
		}

		response.Error(
			c,
			http.StatusForbidden,
			authz.ErrForbidden.Error(),
		)

		c.Abort()
	}
}
