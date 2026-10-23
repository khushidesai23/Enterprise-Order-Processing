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

func (m *webhookRepositoryMock) MarkFailed(ctx context.Context, tx *gorm.DB, payloadID string, processedAt time.Time) error {
	return m.Called(ctx, tx, payloadID, processedAt).Error(0)
}

func (m *webhookRepositoryMock) MarkProcessed(ctx context.Context, tx *gorm.DB, payloadID string, processedAt time.Time) error {
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

func (m *paymentGatewayMock) RefundPayment(ctx context.Context, req GatewayRefundRequest) (*GatewayRefundResponse, error) {
	args := m.Called(ctx, req)
	response, _ := args.Get(0).(*GatewayRefundResponse)
	return response, args.Error(1)
}

func testPaymentService(repo *paymentRepositoryMock, webhookRepo *webhookRepositoryMock, orderService *paymentOrderServiceMock, gateway *paymentGatewayMock) *Service {
	return NewService(repo, webhookRepo, orderService, gateway, "rzp_test_key", zap.NewNop())
}

func TestProcessWebhookRejectsMissingEventID(t *testing.T) {
	ctx := context.Background()
	gateway := new(paymentGatewayMock)
	service := testPaymentService(new(paymentRepositoryMock), new(webhookRepositoryMock), new(paymentOrderServiceMock), gateway)
	body := []byte(`{"event":"payment.captured"}`)
	gateway.On("VerifyWebhookSignature", ctx, body, "sig").Return(nil).Once()

	err := service.ProcessWebhook(ctx, ProcessWebhookRequest{Body: body, Signature: "sig"})

	assert.ErrorIs(t, err, ErrMissingWebhookEvent)
	gateway.AssertExpectations(t)
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

func TestApplyWebhookPaymentStatusFailedKeepsOrderOpenForRetry(t *testing.T) {
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

	err := service.applyWebhookPaymentStatus(ctx, tx, paymentModel, &transactionID, models.PaymentFailed)

	require.NoError(t, err)
	repo.AssertExpectations(t)
	// The order must not be touched: it stays PAYMENT_PENDING with its
	// stock reserved so the customer can retry.
	orderService.AssertNotCalled(t, "RefundOrder", mock.Anything, mock.Anything)
	orderService.AssertNotCalled(t, "MarkOrderPaid", mock.Anything, mock.Anything)
}

func TestApplyWebhookPaymentStatusCaptureAfterFailedAttempt(t *testing.T) {
	ctx := context.Background()
	repo := new(paymentRepositoryMock)
	orderService := new(paymentOrderServiceMock)
	service := testPaymentService(repo, new(webhookRepositoryMock), orderService, new(paymentGatewayMock))
	tx := &gorm.DB{}
	transactionID := "pay_second_attempt"
	paymentModel := &models.Payment{
		BaseModel: models.BaseModel{ID: uuid.New()},
		OrderID:   uuid.New(),
		Status:    models.PaymentFailed,
	}

	repo.On("CompletePayment", ctx, tx, paymentModel.ID, &transactionID, models.PaymentSuccess).Return(nil).Once()
	orderService.On("MarkOrderPaid", tx, paymentModel.OrderID).Return(nil).Once()

	err := service.applyWebhookPaymentStatus(ctx, tx, paymentModel, &transactionID, models.PaymentSuccess)

	require.NoError(t, err)
	repo.AssertExpectations(t)
	orderService.AssertExpectations(t)
}

// Out-of-order and duplicate deliveries must not move a payment backwards.
func TestApplyWebhookPaymentStatusIgnoresStaleEvents(t *testing.T) {
	tests := []struct {
		name    string
		current models.PaymentStatus
		event   models.PaymentStatus
	}{
		{name: "failure after capture", current: models.PaymentSuccess, event: models.PaymentFailed},
		{name: "failure after refund", current: models.PaymentRefunded, event: models.PaymentFailed},
		{name: "capture after refund", current: models.PaymentRefunded, event: models.PaymentSuccess},
		{name: "duplicate capture", current: models.PaymentSuccess, event: models.PaymentSuccess},
		{name: "duplicate failure", current: models.PaymentFailed, event: models.PaymentFailed},
		{name: "duplicate refund", current: models.PaymentRefunded, event: models.PaymentRefunded},
		{name: "authorization", current: models.PaymentPending, event: models.PaymentPending},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repo := new(paymentRepositoryMock)
			orderService := new(paymentOrderServiceMock)
			service := testPaymentService(repo, new(webhookRepositoryMock), orderService, new(paymentGatewayMock))
			paymentModel := &models.Payment{
				BaseModel: models.BaseModel{ID: uuid.New()},
				OrderID:   uuid.New(),
				Status:    test.current,
			}
			transactionID := "pay_x"

			err := service.applyWebhookPaymentStatus(context.Background(), &gorm.DB{}, paymentModel, &transactionID, test.event)

			require.NoError(t, err)
			// No expectations were set: any repository or order call fails the mock.
			repo.AssertExpectations(t)
			orderService.AssertExpectations(t)
		})
	}
}

func TestApplyWebhookPaymentStatusCaptureForCancelledOrderIsAcknowledged(t *testing.T) {
	ctx := context.Background()
	repo := new(paymentRepositoryMock)
	orderService := new(paymentOrderServiceMock)
	service := testPaymentService(repo, new(webhookRepositoryMock), orderService, new(paymentGatewayMock))
	tx := &gorm.DB{}
	transactionID := "pay_late"
	paymentModel := &models.Payment{
		BaseModel: models.BaseModel{ID: uuid.New()},
		OrderID:   uuid.New(),
		Status:    models.PaymentPending,
	}

	repo.On("CompletePayment", ctx, tx, paymentModel.ID, &transactionID, models.PaymentSuccess).Return(nil).Once()
	orderService.On("MarkOrderPaid", tx, paymentModel.OrderID).Return(ordermodule.ErrOrderAlreadyCancelled).Once()

	err := service.applyWebhookPaymentStatus(ctx, tx, paymentModel, &transactionID, models.PaymentSuccess)

	// Recorded as captured (so it can be refunded) and acknowledged.
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestProcessWebhookAcknowledgesUnhandledEvents(t *testing.T) {
	for name, body := range map[string]string{
		"order.paid":       `{"event":"order.paid"}`,
		"refund.processed": `{"event":"refund.processed"}`,
		"partial refund":   `{"event":"refund.created","payload":{"payment":{"entity":{"id":"pay_1","order_id":"order_1","refund_status":"partial"}}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			gateway := new(paymentGatewayMock)
			repo := new(paymentRepositoryMock)
			service := testPaymentService(repo, new(webhookRepositoryMock), new(paymentOrderServiceMock), gateway)
			gateway.On("VerifyWebhookSignature", ctx, []byte(body), "sig").Return(nil).Once()

			err := service.ProcessWebhook(ctx, ProcessWebhookRequest{Body: []byte(body), Signature: "sig", EventID: "evt_1"})

			require.NoError(t, err)
			repo.AssertNotCalled(t, "Begin", mock.Anything)
		})
	}
}

func TestToPaiseRounds(t *testing.T) {
	assert.Equal(t, 1999, toPaise(19.99))
	assert.Equal(t, 149950, toPaise(1499.5))
	assert.Equal(t, 1, toPaise(0.01))
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
