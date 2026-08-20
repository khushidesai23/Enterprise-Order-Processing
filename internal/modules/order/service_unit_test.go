package order

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/test/mocks"
)

type orderRepositoryMock struct {
	mock.Mock
}

func (m *orderRepositoryMock) Begin(ctx context.Context) *gorm.DB {
	args := m.Called(ctx)
	db, _ := args.Get(0).(*gorm.DB)
	return db
}

func (m *orderRepositoryMock) Create(ctx context.Context, tx *gorm.DB, order *models.Order) error {
	return m.Called(ctx, tx, order).Error(0)
}

func (m *orderRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	args := m.Called(ctx, id)
	order, _ := args.Get(0).(*models.Order)
	return order, args.Error(1)
}

func (m *orderRepositoryMock) GetByIDTx(ctx context.Context, tx *gorm.DB, id uuid.UUID) (*models.Order, error) {
	args := m.Called(ctx, tx, id)
	order, _ := args.Get(0).(*models.Order)
	return order, args.Error(1)
}

func (m *orderRepositoryMock) GetAll(ctx context.Context) ([]models.Order, error) {
	args := m.Called(ctx)
	orders, _ := args.Get(0).([]models.Order)
	return orders, args.Error(1)
}

func (m *orderRepositoryMock) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	args := m.Called(ctx, userID)
	orders, _ := args.Get(0).([]models.Order)
	return orders, args.Error(1)
}

func (m *orderRepositoryMock) UpdateTotalAmount(ctx context.Context, tx *gorm.DB, id uuid.UUID, total float64) error {
	return m.Called(ctx, tx, id, total).Error(0)
}

func (m *orderRepositoryMock) UpdateStatusTx(ctx context.Context, tx *gorm.DB, id uuid.UUID, status models.OrderStatus) error {
	return m.Called(ctx, tx, id, status).Error(0)
}

type orderItemRepositoryMock struct {
	mock.Mock
}

func (m *orderItemRepositoryMock) CreateMany(ctx context.Context, tx *gorm.DB, items []models.OrderItem) error {
	return m.Called(ctx, tx, items).Error(0)
}

type orderProductRepositoryMock struct {
	mock.Mock
}

func (m *orderProductRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	args := m.Called(ctx, id)
	product, _ := args.Get(0).(*models.Product)
	return product, args.Error(1)
}

type orderInventoryRepositoryMock struct {
	mock.Mock
}

func (m *orderInventoryRepositoryMock) GetByProductIDTx(ctx context.Context, tx *gorm.DB, productID uuid.UUID) (*models.Inventory, error) {
	args := m.Called(ctx, tx, productID)
	inventory, _ := args.Get(0).(*models.Inventory)
	return inventory, args.Error(1)
}

func (m *orderInventoryRepositoryMock) ReserveStockTx(ctx context.Context, tx *gorm.DB, productID uuid.UUID, quantity int) error {
	return m.Called(ctx, tx, productID, quantity).Error(0)
}

func (m *orderInventoryRepositoryMock) ReleaseReservedStockTx(ctx context.Context, tx *gorm.DB, productID uuid.UUID, quantity int) error {
	return m.Called(ctx, tx, productID, quantity).Error(0)
}

func (m *orderInventoryRepositoryMock) ConfirmReservedStockTx(ctx context.Context, tx *gorm.DB, productID uuid.UUID, quantity int) error {
	return m.Called(ctx, tx, productID, quantity).Error(0)
}

func testOrderService(orderRepo *orderRepositoryMock, userRepo *mocks.UserRepositoryMock, productRepo *orderProductRepositoryMock, inventoryRepo *orderInventoryRepositoryMock) *Service {
	return NewService(orderRepo, new(orderItemRepositoryMock), userRepo, productRepo, inventoryRepo, zap.NewNop())
}

func TestEnsureUniqueProducts(t *testing.T) {
	service := testOrderService(new(orderRepositoryMock), new(mocks.UserRepositoryMock), new(orderProductRepositoryMock), new(orderInventoryRepositoryMock))
	productID := uuid.New()

	err := service.ensureUniqueProducts([]CreateOrderItemRequest{
		{ProductID: productID, Quantity: 1},
		{ProductID: productID, Quantity: 2},
	})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrDuplicateProduct)
}

func TestValidateStatusTransition(t *testing.T) {
	service := testOrderService(new(orderRepositoryMock), new(mocks.UserRepositoryMock), new(orderProductRepositoryMock), new(orderInventoryRepositoryMock))

	require.NoError(t, service.validateStatusTransition(models.OrderCreated, models.OrderPaymentPending))
	assert.ErrorIs(t, service.validateStatusTransition(models.OrderDelivered, models.OrderShipped), ErrOrderAlreadyCompleted)
	assert.ErrorIs(t, service.validateStatusTransition(models.OrderPaid, models.OrderDelivered), ErrInvalidOrderStatus)
}

func TestConfirmReservedInventory(t *testing.T) {
	ctx := context.Background()
	orderRepo := new(orderRepositoryMock)
	inventoryRepo := new(orderInventoryRepositoryMock)
	service := testOrderService(orderRepo, new(mocks.UserRepositoryMock), new(orderProductRepositoryMock), inventoryRepo)
	tx := &gorm.DB{}
	productID := uuid.New()
	orderModel := &models.Order{
		BaseModel: models.BaseModel{ID: uuid.New()},
		Items: []models.OrderItem{
			{ProductID: productID, Quantity: 2},
		},
	}

	inventoryRepo.On("GetByProductIDTx", ctx, tx, productID).Return(&models.Inventory{
		ProductID:        productID,
		ReservedQuantity: 3,
	}, nil).Once()
	inventoryRepo.On("ConfirmReservedStockTx", ctx, tx, productID, 2).Return(nil).Once()

	err := service.confirmReservedInventory(ctx, tx, orderModel)

	require.NoError(t, err)
	inventoryRepo.AssertExpectations(t)
}

func TestReleaseReservedInventoryRejectsInsufficientReserved(t *testing.T) {
	ctx := context.Background()
	inventoryRepo := new(orderInventoryRepositoryMock)
	service := testOrderService(new(orderRepositoryMock), new(mocks.UserRepositoryMock), new(orderProductRepositoryMock), inventoryRepo)
	tx := &gorm.DB{}
	productID := uuid.New()
	orderModel := &models.Order{
		Items: []models.OrderItem{
			{ProductID: productID, Quantity: 5},
		},
	}

	inventoryRepo.On("GetByProductIDTx", ctx, tx, productID).Return(&models.Inventory{
		ProductID:        productID,
		ReservedQuantity: 2,
	}, nil).Once()

	err := service.releaseReservedInventory(ctx, tx, orderModel)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInsufficientReserved)
}

func TestGetOrdersByUser(t *testing.T) {
	ctx := context.Background()
	orderRepo := new(orderRepositoryMock)
	userRepo := new(mocks.UserRepositoryMock)
	service := testOrderService(orderRepo, userRepo, new(orderProductRepositoryMock), new(orderInventoryRepositoryMock))
	userID := uuid.New()

	userRepo.On("GetByID", ctx, userID).Return(nil, nil).Once()

	orders, err := service.GetOrdersByUser(ctx, userID)

	require.Error(t, err)
	assert.Nil(t, orders)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func TestTransitionOrderUpdatesStatus(t *testing.T) {
	ctx := context.Background()
	orderRepo := new(orderRepositoryMock)
	service := testOrderService(orderRepo, new(mocks.UserRepositoryMock), new(orderProductRepositoryMock), new(orderInventoryRepositoryMock))
	tx := &gorm.DB{}
	orderModel := &models.Order{
		BaseModel: models.BaseModel{ID: uuid.New()},
		Status:    models.OrderPaid,
		UserID:    uuid.New(),
	}

	orderRepo.On("UpdateStatusTx", ctx, tx, orderModel.ID, models.OrderPacked).Return(nil).Once()

	err := service.transitionOrder(ctx, tx, orderModel, models.OrderPacked)

	require.NoError(t, err)
	orderRepo.AssertExpectations(t)
}
