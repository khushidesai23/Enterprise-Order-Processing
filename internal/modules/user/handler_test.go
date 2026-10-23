package user

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/authz"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

// fakeService records which user the handler acted on.
type fakeService struct {
	Service
	called uuid.UUID
}

func (f *fakeService) GetByID(_ context.Context, id uuid.UUID) (*UserResponse, error) {
	f.called = id
	return &UserResponse{ID: id.String()}, nil
}

func (f *fakeService) Update(_ context.Context, id uuid.UUID, _ UpdateUserRequest) (*UserResponse, error) {
	f.called = id
	return &UserResponse{ID: id.String()}, nil
}

func (f *fakeService) UpdateRole(_ context.Context, id uuid.UUID, role models.Role) (*UserResponse, error) {
	f.called = id
	return &UserResponse{ID: id.String(), Role: role}, nil
}

func newUserTestRouter(service Service, principal authz.Principal) *gin.Engine {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(service)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		authz.SetPrincipal(c, principal)
		c.Next()
	})
	router.GET("/users/:id", handler.GetByID)
	router.PUT("/users/:id", handler.Update)
	router.PATCH("/users/:id/role", handler.UpdateRole)
	return router
}

func send(router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestUserHandlerOwnership(t *testing.T) {
	self := uuid.New()
	other := uuid.New()
	customer := authz.Principal{UserID: self, Role: models.RoleCustomer}
	admin := authz.Principal{UserID: uuid.New(), Role: models.RoleAdmin}

	tests := []struct {
		name       string
		principal  authz.Principal
		method     string
		target     uuid.UUID
		body       string
		wantStatus int
	}{
		{name: "customer reads self", principal: customer, method: http.MethodGet, target: self, wantStatus: http.StatusOK},
		{name: "customer reads other", principal: customer, method: http.MethodGet, target: other, wantStatus: http.StatusNotFound},
		{name: "admin reads other", principal: admin, method: http.MethodGet, target: other, wantStatus: http.StatusOK},
		{name: "customer updates self", principal: customer, method: http.MethodPut, target: self, body: `{"first_name":"New"}`, wantStatus: http.StatusOK},
		{name: "customer updates other", principal: customer, method: http.MethodPut, target: other, body: `{"first_name":"New"}`, wantStatus: http.StatusNotFound},
		{name: "customer cannot change is_active", principal: customer, method: http.MethodPut, target: self, body: `{"is_active":true}`, wantStatus: http.StatusForbidden},
		{name: "admin changes is_active", principal: admin, method: http.MethodPut, target: other, body: `{"is_active":false}`, wantStatus: http.StatusOK},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &fakeService{}
			router := newUserTestRouter(service, test.principal)

			recorder := send(router, test.method, "/users/"+test.target.String(), test.body)

			require.Equal(t, test.wantStatus, recorder.Code, recorder.Body.String())
			if test.wantStatus == http.StatusOK {
				assert.Equal(t, test.target, service.called)
			} else {
				assert.Equal(t, uuid.Nil, service.called, "service must not be reached")
			}
		})
	}
}

func TestUpdateRole(t *testing.T) {
	admin := authz.Principal{UserID: uuid.New(), Role: models.RoleAdmin}
	target := uuid.New()

	t.Run("promotes another user", func(t *testing.T) {
		service := &fakeService{}
		recorder := send(newUserTestRouter(service, admin), http.MethodPatch, "/users/"+target.String()+"/role", `{"role":"admin"}`)

		require.Equal(t, http.StatusOK, recorder.Code, recorder.Body.String())
		assert.Equal(t, target, service.called)
	})

	t.Run("refuses to change own role", func(t *testing.T) {
		service := &fakeService{}
		recorder := send(newUserTestRouter(service, admin), http.MethodPatch, "/users/"+admin.UserID.String()+"/role", `{"role":"customer"}`)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Equal(t, uuid.Nil, service.called)
	})
}

func TestServiceUpdateRoleRejectsUnknownRole(t *testing.T) {
	service := NewService(nil, nil)

	_, err := service.UpdateRole(context.Background(), uuid.New(), models.Role("superuser"))

	require.ErrorIs(t, err, ErrInvalidRole)
}
