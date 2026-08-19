package payment

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	ordermodule "github.com/khushidesai23/Enterprise-Order-Processing/internal/modules/order"
)

type paymentRepositoryMock struct {
	mock.Mock
}

func (m *paymentRepositoryMock) Begin(ctx context.Context) *gorm.DB {
	args := m.Called(ctx)
	db, _ := args.Get(0).(*gorm.DB)
	return db
}

func (m *paymentRepositoryMock) Create(ctx context.Context, tx *gorm.DB, payment *models.Payment) error {
	return m.Called(ctx, tx, payment).Error(0)
}

func (m *paymentRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*models.Payment, error) {
	args := m.Called(ctx, id)
	payment, _ := args.Get(0).(*models.Payment)
	return payment, args.Error(1)
}

func (m *paymentRepositoryMock) GetByIDTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*models.Payment, error) {
	args := m.Called(ctx, tx, id)
	payment, _ := args.Get(0).(*models.Payment)
	return payment, args.Error(1)
}

func (m *paymentRepositoryMock) GetByOrderID(ctx context.Context, orderID uuid.UUID) (*models.Payment, error) {
	args := m.Called(ctx, orderID)
	payment, _ := args.Get(0).(*models.Payment)
	return payment, args.Error(1)
}

func (m *paymentRepositoryMock) GetByOrderIDTx(ctx context.Context, tx *gorm.DB, orderID uuid.UUID) (*models.Payment, error) {
	args := m.Called(ctx, tx, orderID)
	payment, _ := args.Get(0).(*models.Payment)
	return payment, args.Error(1)
}

func (m *paymentRepositoryMock) GetByGatewayOrderIDTx(ctx context.Context, tx *gorm.DB, gatewayOrderID string) (*models.Payment, error) {
	args := m.Called(ctx, tx, gatewayOrderID)
	payment, _ := args.Get(0).(*models.Payment)
	return payment, args.Error(1)
}

func (m *paymentRepositoryMock) GetAll(ctx context.Context) ([]models.Payment, error) {
	args := m.Called(ctx)
	payments, _ := args.Get(0).([]models.Payment)
	return payments, args.Error(1)
}

func (m *paymentRepositoryMock) UpdateStatus(ctx context.Context, tx *gorm.DB, id uuid.UUID, status models.PaymentStatus) error {
	return m.Called(ctx, tx, id, status).Error(0)
}

func (m *paymentRepositoryMock) CompletePayment(ctx context.Context, tx *gorm.DB, id uuid.UUID, transactionID *string, status models.PaymentStatus) error {
	return m.Called(ctx, tx, id, transactionID, status).Error(0)
}

type webhookRepositoryMock struct {
	mock.Mock
}

func (m *webhookRepositoryMock) Create(ctx context.Context, tx *gorm.DB, webhook *models.PaymentWebhook) error {
	return m.Called(ctx, tx, webhook).Error(0)
}

func (m *webhookRepositoryMock) MarkProcessed(ctx context.Context, tx *gorm.DB, payloadID string, processedAt time.Time) error {
	return m.Called(ctx, tx, payloadID, processedAt).Error(0)
}

func (m *webhookRepositoryMock) MarkFailed(ctx context.Context, tx *gorm.DB, payloadID string, processedAt time.Time) error {
	return m.Called(ctx, tx, payloadID, processedAt).Error(0)
}

type paymentOrderServiceMock struct {
	mock.Mock
}

func (m *paymentOrderServiceMock) GetOrderForPaymentTx(tx *gorm.DB, id uuid.UUID) (*models.Order, error) {
	args := m.Called(tx, id)
	order, _ := args.Get(0).(*models.Order)
	return order, args.Error(1)
}

func (m *paymentOrderServiceMock) GetOrderForPayment(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	args := m.Called(ctx, id)
	order, _ := args.Get(0).(*models.Order)
	return order, args.Error(1)
}

func (m *paymentOrderServiceMock) MarkOrderPaymentPending(tx *gorm.DB, orderID uuid.UUID) error {
	return m.Called(tx, orderID).Error(0)
}

func (m *paymentOrderServiceMock) MarkOrderPaid(tx *gorm.DB, orderID uuid.UUID) error {
	return m.Called(tx, orderID).Error(0)
}

func (m *paymentOrderServiceMock) CancelOrderByPaymentFailure(tx *gorm.DB, orderID uuid.UUID) error {
	return m.Called(tx, orderID).Error(0)
}

func (m *paymentOrderServiceMock) RefundOrder(tx *gorm.DB, orderID uuid.UUID) error {
	return m.Called(tx, orderID).Error(0)
}

type paymentGatewayMock struct {
	mock.Mock
}

func (m *paymentGatewayMock) CreateOrder(ctx context.Context, req GatewayOrderRequest) (*GatewayOrderResponse, error) {
	args := m.Called(ctx, req)
	response, _ := args.Get(0).(*GatewayOrderResponse)
	return response, args.Error(1)
}

func (m *paymentGatewayMock) VerifyCheckoutSignature(ctx context.Context, orderID string, paymentID string, signature string) error {
	return m.Called(ctx, orderID, paymentID, signature).Error(0)
}

func (m *paymentGatewayMock) VerifyWebhookSignature(ctx context.Context, body []byte, signature string) error {
	return m.Called(ctx, body, signature).Error(0)
}

func testPaymentService(repo *paymentRepositoryMock, webhookRepo *webhookRepositoryMock, orderService *paymentOrderServiceMock, gateway *paymentGatewayMock) *Service {
	return NewService(repo, webhookRepo, orderService, gateway, "rzp_test_key", zap.NewNop())
}

func TestValidateOrderForPayment(t *testing.T) {
	service := testPaymentService(new(paymentRepositoryMock), new(webhookRepositoryMock), new(paymentOrderServiceMock), new(paymentGatewayMock))

	require.NoError(t, service.validateOrderForPayment(models.OrderCreated))
	require.NoError(t, service.validateOrderForPayment(models.OrderPaymentPending))
	assert.ErrorIs(t, service.validateOrderForPayment(models.OrderPaid), ErrOrderAlreadyPaid)
	assert.ErrorIs(t, service.validateOrderForPayment(models.OrderCancelled), ErrOrderCancelled)
	assert.ErrorIs(t, service.validateOrderForPayment(models.OrderDelivered), ErrOrderNotEligibleForPayment)
}

func TestApplyWebhookPaymentStatusSuccess(t *testing.T) {
	ctx := context.Background()
	repo := new(paymentRepositoryMock)
	orderService := new(paymentOrderServiceMock)
	service := testPaymentService(repo, new(webhookRepositoryMock), orderService, new(paymentGatewayMock))
	tx := &gorm.DB{}
	transactionID := "txn_1"
	paymentModel := &models.Payment{
		BaseModel: models.BaseModel{ID: uuid.New()},
		OrderID:   uuid.New(),
		Status:    models.PaymentPending,
	}

	repo.On("CompletePayment", ctx, tx, paymentModel.ID, &transactionID, models.PaymentSuccess).Return(nil).Once()
	orderService.On("MarkOrderPaid", tx, paymentModel.OrderID).Return(nil).Once()

	err := service.applyWebhookPaymentStatus(ctx, tx, paymentModel, &transactionID, models.PaymentSuccess)

	require.NoError(t, err)
}

func TestApplyWebhookPaymentStatusFailedCancelsOrder(t *testing.T) {
	ctx := context.Background()
	repo := new(paymentRepositoryMock)
	orderService := new(paymentOrderServiceMock)
	service := testPaymentService(repo, new(webhookRepositoryMock), orderService, new(paymentGatewayMock))
	tx := &gorm.DB{}
	transactionID := "txn_2"
	paymentModel := &models.Payment{
		BaseModel: models.BaseModel{ID: uuid.New()},
		OrderID:   uuid.New(),
		Status:    models.PaymentPending,
	}

	repo.On("CompletePayment", ctx, tx, paymentModel.ID, &transactionID, models.PaymentFailed).Return(nil).Once()
	orderService.On("CancelOrderByPaymentFailure", tx, paymentModel.OrderID).Return(nil).Once()

	err := service.applyWebhookPaymentStatus(ctx, tx, paymentModel, &transactionID, models.PaymentFailed)

	require.NoError(t, err)
}

func TestApplyWebhookPaymentStatusRefunded(t *testing.T) {
	ctx := context.Background()
	repo := new(paymentRepositoryMock)
	orderService := new(paymentOrderServiceMock)
	service := testPaymentService(repo, new(webhookRepositoryMock), orderService, new(paymentGatewayMock))
	tx := &gorm.DB{}
	paymentModel := &models.Payment{
		BaseModel: models.BaseModel{ID: uuid.New()},
		OrderID:   uuid.New(),
		Status:    models.PaymentSuccess,
	}

	repo.On("UpdateStatus", ctx, tx, paymentModel.ID, models.PaymentRefunded).Return(nil).Once()
	orderService.On("RefundOrder", tx, paymentModel.OrderID).Return(nil).Once()

	err := service.applyWebhookPaymentStatus(ctx, tx, paymentModel, nil, models.PaymentRefunded)

	require.NoError(t, err)
}

func TestMapOrderError(t *testing.T) {
	assert.ErrorIs(t, mapOrderError(ordermodule.ErrOrderNotFound), ErrOrderNotFound)
	assert.ErrorIs(t, mapOrderError(ordermodule.ErrOrderAlreadyCancelled), ErrOrderCancelled)
	assert.ErrorIs(t, mapOrderError(ordermodule.ErrInvalidOrderStatus), ErrOrderNotEligibleForPayment)

	original := errors.New("boom")
	assert.ErrorIs(t, mapOrderError(original), original)
}

func TestDuplicateKeyAndStringHelpers(t *testing.T) {
	assert.True(t, isDuplicateKey(errors.New("duplicate key value violates unique constraint")))
	assert.False(t, isDuplicateKey(errors.New("random error")))
	assert.Equal(t, "", stringValue(nil))

	value := "abc"
	assert.Equal(t, "abc", stringValue(&value))
}
