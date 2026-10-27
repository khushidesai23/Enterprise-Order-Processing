package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/api/handlers"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/auth"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/category"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/inventory"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/order"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/payment"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/product"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/review"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/user"
)

// adminOnlyRoutes must reject customers before reaching a handler. The
// handlers below have no services, so a request that got past the role
// check would panic instead of returning 403.
var adminOnlyRoutes = []struct {
	method string
	path   string
}{
	{http.MethodGet, "/api/v1/users"},
	{http.MethodDelete, "/api/v1/users/{id}"},
	{http.MethodPatch, "/api/v1/users/{id}/role"},
	{http.MethodPost, "/api/v1/categories"},
	{http.MethodPut, "/api/v1/categories/{id}"},
	{http.MethodDelete, "/api/v1/categories/{id}"},
	{http.MethodPost, "/api/v1/products"},
	{http.MethodPut, "/api/v1/products/{id}"},
	{http.MethodDelete, "/api/v1/products/{id}"},
	{http.MethodPost, "/api/v1/inventory"},
	{http.MethodPut, "/api/v1/inventory/{id}"},
	{http.MethodPatch, "/api/v1/inventory/{id}/add-stock"},
	{http.MethodPatch, "/api/v1/inventory/{id}/remove-stock"},
	{http.MethodPatch, "/api/v1/inventory/{id}/reserve"},
	{http.MethodPatch, "/api/v1/inventory/{id}/release"},
	{http.MethodPatch, "/api/v1/inventory/{id}/confirm"},
	{http.MethodPatch, "/api/v1/orders/{id}/status"},
	{http.MethodGet, "/api/v1/orders/summary"},
	{http.MethodGet, "/api/v1/reviews"},
	{http.MethodGet, "/api/v1/payments"},
	{http.MethodGet, "/api/v1/payments/summary"},
	{http.MethodPost, "/api/v1/payments/{id}/refund"},
}

func newTestRouter(jwtManager *auth.JWTManager) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	Register(
		router,
		handlers.NewHealthHandler(nil, nil),
		user.NewHandler(nil),
		product.NewHandler(nil),
		category.NewHandler(nil),
		inventory.NewHandler(nil),
		order.NewHandler(nil),
		payment.NewHandler(nil),
		review.NewHandler(nil),
		auth.NewHandler(nil),
		jwtManager,
	)
	return router
}

func TestAdminOnlyRoutes(t *testing.T) {
	jwtManager := auth.NewJWTManager("secret", time.Hour)
	router := newTestRouter(jwtManager)
	customerToken, err := jwtManager.GenerateToken(uuid.NewString(), "customer@example.com", "customer")
	require.NoError(t, err)

	for _, route := range adminOnlyRoutes {
		path := strings.ReplaceAll(route.path, "{id}", uuid.NewString())

		t.Run(route.method+" "+route.path, func(t *testing.T) {
			for name, test := range map[string]struct {
				token string
				want  int
			}{
				"anonymous": {token: "", want: http.StatusUnauthorized},
				"customer":  {token: customerToken, want: http.StatusForbidden},
			} {
				request := httptest.NewRequest(route.method, path, strings.NewReader("{}"))
				request.Header.Set("Content-Type", "application/json")
				if test.token != "" {
					request.Header.Set("Authorization", "Bearer "+test.token)
				}
				recorder := httptest.NewRecorder()

				router.ServeHTTP(recorder, request)

				assert.Equal(t, test.want, recorder.Code, "%s: %s", name, recorder.Body.String())
			}
		})
	}
}

// TestEveryMutatingRouteIsAuthenticated guards against a new write route
// being registered without the auth middleware.
func TestEveryMutatingRouteIsAuthenticated(t *testing.T) {
	router := newTestRouter(auth.NewJWTManager("secret", time.Hour))
	public := map[string]bool{
		"POST /api/v1/auth/login":       true,
		"POST /api/v1/users":            true, // sign-up
		"POST /api/v1/payments/webhook": true, // verified by signature
	}

	for _, route := range router.Routes() {
		if route.Method == http.MethodGet || public[route.Method+" "+route.Path] {
			continue
		}

		path := route.Path
		for _, param := range []string{":id", ":productId", ":orderId", ":categoryId"} {
			path = strings.ReplaceAll(path, param, uuid.NewString())
		}
		request := httptest.NewRequest(route.Method, path, strings.NewReader("{}"))
		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, request)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code, "%s %s", route.Method, route.Path)
	}
}
