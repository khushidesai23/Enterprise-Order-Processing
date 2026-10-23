package order

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/authz"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/test/mocks"
	"github.com/stretchr/testify/require"
)

// withPrincipal stands in for the auth middleware.
func withPrincipal(principal authz.Principal) gin.HandlerFunc {
	return func(c *gin.Context) {
		authz.SetPrincipal(c, principal)
		c.Next()
	}
}

func TestGetOrdersPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	admin := authz.Principal{UserID: uuid.New(), Role: models.RoleAdmin}

	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantLimit  int
		wantOffset int
	}{
		{name: "default page", wantStatus: http.StatusOK, wantLimit: DefaultOrderPageSize, wantOffset: 0},
		{name: "requested page", query: "?page=3&limit=25", wantStatus: http.StatusOK, wantLimit: 25, wantOffset: 50},
		{name: "rejects invalid page", query: "?page=0", wantStatus: http.StatusBadRequest},
		{name: "rejects oversized limit", query: "?limit=101", wantStatus: http.StatusBadRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := new(orderRepositoryMock)
			if test.wantStatus == http.StatusOK {
				repo.On("GetAll", ctx, test.wantLimit, test.wantOffset).Return([]models.Order{}, nil).Once()
			}
			service := testOrderService(repo, new(mocks.UserRepositoryMock), new(orderProductRepositoryMock), new(orderInventoryRepositoryMock))
			router := gin.New()
			router.GET("/orders", withPrincipal(admin), NewHandler(service).GetOrders)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/orders"+test.query, nil)

			router.ServeHTTP(recorder, request)

			require.Equal(t, test.wantStatus, recorder.Code)
			repo.AssertExpectations(t)
		})
	}
}

func TestGetOrdersScopesCustomersToTheirOwnOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()
	customer := authz.Principal{UserID: uuid.New(), Role: models.RoleCustomer}

	repo := new(orderRepositoryMock)
	userRepo := new(mocks.UserRepositoryMock)
	userRepo.On("GetByID", ctx, customer.UserID).Return(&models.User{}, nil).Once()
	repo.On("GetByUserIDPage", ctx, customer.UserID, DefaultOrderPageSize, 0).Return([]models.Order{}, nil).Once()
	service := testOrderService(repo, userRepo, new(orderProductRepositoryMock), new(orderInventoryRepositoryMock))
	router := gin.New()
	router.GET("/orders", withPrincipal(customer), NewHandler(service).GetOrders)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/orders", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	repo.AssertExpectations(t)
	repo.AssertNotCalled(t, "GetAll")
}

func TestGetOrdersRequiresAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := new(orderRepositoryMock)
	service := testOrderService(repo, new(mocks.UserRepositoryMock), new(orderProductRepositoryMock), new(orderInventoryRepositoryMock))
	router := gin.New()
	router.GET("/orders", NewHandler(service).GetOrders)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/orders", nil))

	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	repo.AssertNotCalled(t, "GetAll")
}

func TestGetOrderEnforcesOwnership(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ownerID := uuid.New()

	tests := []struct {
		name       string
		principal  authz.Principal
		wantStatus int
	}{
		{name: "owner", principal: authz.Principal{UserID: ownerID, Role: models.RoleCustomer}, wantStatus: http.StatusOK},
		{name: "admin", principal: authz.Principal{UserID: uuid.New(), Role: models.RoleAdmin}, wantStatus: http.StatusOK},
		// Not 403: another user's order must look the same as a missing one.
		{name: "other customer", principal: authz.Principal{UserID: uuid.New(), Role: models.RoleCustomer}, wantStatus: http.StatusNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			orderID := uuid.New()
			repo := new(orderRepositoryMock)
			repo.On("GetByID", context.Background(), orderID).Return(&models.Order{
				BaseModel: models.BaseModel{ID: orderID},
				UserID:    ownerID,
				Status:    models.OrderCreated,
			}, nil).Once()
			service := testOrderService(repo, new(mocks.UserRepositoryMock), new(orderProductRepositoryMock), new(orderInventoryRepositoryMock))
			router := gin.New()
			router.GET("/orders/:id", withPrincipal(test.principal), NewHandler(service).GetOrder)
			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/orders/"+orderID.String(), nil))

			require.Equal(t, test.wantStatus, recorder.Code, recorder.Body.String())
			repo.AssertExpectations(t)
		})
	}
}
