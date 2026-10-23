package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/authz"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/auth"
)

func newAuthTestRouter(jwtManager *auth.JWTManager, captured *authz.Principal) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	authMiddleware := AuthMiddleware(jwtManager)
	router.GET("/whoami", authMiddleware, func(c *gin.Context) {
		*captured, _ = authz.FromContext(c)
		c.Status(http.StatusOK)
	})
	router.GET("/admin", authMiddleware, RequireRole(models.RoleAdmin), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.GET("/unauthenticated-admin", RequireRole(models.RoleAdmin), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return router
}

func serve(t *testing.T, router http.Handler, path string, token string) int {
	t.Helper()
	request := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder.Code
}

func TestAuthMiddlewareSetsPrincipal(t *testing.T) {
	jwtManager := auth.NewJWTManager("secret", time.Hour)
	userID := uuid.New()

	tests := []struct {
		name     string
		role     string
		wantRole models.Role
	}{
		{name: "admin", role: "admin", wantRole: models.RoleAdmin},
		{name: "customer", role: "customer", wantRole: models.RoleCustomer},
		{name: "token without role is a customer", role: "", wantRole: models.RoleCustomer},
		{name: "unknown role is a customer", role: "superuser", wantRole: models.RoleCustomer},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var principal authz.Principal
			router := newAuthTestRouter(jwtManager, &principal)
			token, err := jwtManager.GenerateToken(userID.String(), "user@example.com", test.role)
			require.NoError(t, err)

			require.Equal(t, http.StatusOK, serve(t, router, "/whoami", token))
			assert.Equal(t, userID, principal.UserID)
			assert.Equal(t, test.wantRole, principal.Role)
		})
	}
}

func TestAuthMiddlewareRejectsNonUUIDSubject(t *testing.T) {
	jwtManager := auth.NewJWTManager("secret", time.Hour)
	var principal authz.Principal
	router := newAuthTestRouter(jwtManager, &principal)
	token, err := jwtManager.GenerateToken("not-a-uuid", "user@example.com", "admin")
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, serve(t, router, "/whoami", token))
}

func TestRequireRole(t *testing.T) {
	jwtManager := auth.NewJWTManager("secret", time.Hour)
	var principal authz.Principal
	router := newAuthTestRouter(jwtManager, &principal)

	adminToken, err := jwtManager.GenerateToken(uuid.NewString(), "admin@example.com", "admin")
	require.NoError(t, err)
	customerToken, err := jwtManager.GenerateToken(uuid.NewString(), "user@example.com", "customer")
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, serve(t, router, "/admin", adminToken))
	assert.Equal(t, http.StatusForbidden, serve(t, router, "/admin", customerToken))
	assert.Equal(t, http.StatusUnauthorized, serve(t, router, "/admin", ""))
	assert.Equal(t, http.StatusUnauthorized, serve(t, router, "/unauthenticated-admin", ""))
}
