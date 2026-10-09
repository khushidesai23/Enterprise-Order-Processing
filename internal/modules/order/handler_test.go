package order

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/test/mocks"
	"github.com/stretchr/testify/require"
)

func TestGetOrdersPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx := context.Background()

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
			router.GET("/orders", NewHandler(service).GetOrders)
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/orders"+test.query, nil)

			router.ServeHTTP(recorder, request)

			require.Equal(t, test.wantStatus, recorder.Code)
			repo.AssertExpectations(t)
		})
	}
}
