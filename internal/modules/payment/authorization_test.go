package payment

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/authz"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

func TestGetPaymentEnforcesOrderOwnership(t *testing.T) {
	ctx := context.Background()
	ownerID := uuid.New()

	tests := []struct {
		name      string
		principal authz.Principal
		wantErr   error
	}{
		{name: "owner", principal: authz.Principal{UserID: ownerID, Role: models.RoleCustomer}},
		{name: "admin", principal: authz.Principal{UserID: uuid.New(), Role: models.RoleAdmin}},
		{name: "other customer", principal: authz.Principal{UserID: uuid.New(), Role: models.RoleCustomer}, wantErr: ErrPaymentNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := new(paymentRepositoryMock)
			orderService := new(paymentOrderServiceMock)
			service := testPaymentService(repo, new(webhookRepositoryMock), orderService, new(paymentGatewayMock))
			paymentModel := &models.Payment{BaseModel: models.BaseModel{ID: uuid.New()}, OrderID: uuid.New()}

			repo.On("GetByID", ctx, paymentModel.ID).Return(paymentModel, nil).Once()
			if !test.principal.IsAdmin() {
				orderService.On("GetOrderForPayment", ctx, paymentModel.OrderID).
					Return(&models.Order{UserID: ownerID}, nil).Once()
			}

			got, err := service.GetPayment(ctx, paymentModel.ID, test.principal)

			if test.wantErr != nil {
				require.ErrorIs(t, err, test.wantErr)
				assert.Nil(t, got)
			} else {
				require.NoError(t, err)
				assert.Equal(t, paymentModel.ID, got.ID)
			}
			repo.AssertExpectations(t)
			orderService.AssertExpectations(t)
		})
	}
}

func TestGetPaymentByOrderEnforcesOrderOwnership(t *testing.T) {
	ctx := context.Background()
	repo := new(paymentRepositoryMock)
	orderService := new(paymentOrderServiceMock)
	service := testPaymentService(repo, new(webhookRepositoryMock), orderService, new(paymentGatewayMock))
	orderID := uuid.New()
	orderService.On("GetOrderForPayment", ctx, orderID).Return(&models.Order{UserID: uuid.New()}, nil).Once()

	got, err := service.GetPaymentByOrder(
		ctx,
		orderID,
		authz.Principal{UserID: uuid.New(), Role: models.RoleCustomer},
	)

	require.ErrorIs(t, err, ErrOrderNotFound)
	assert.Nil(t, got)
	repo.AssertNotCalled(t, "GetByOrderID")
}
