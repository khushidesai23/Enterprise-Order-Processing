package inventory

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
)

type inventoryRepositoryMock struct {
	mock.Mock
}

func (m *inventoryRepositoryMock) Create(ctx context.Context, inventory *models.Inventory) error {
	return m.Called(ctx, inventory).Error(0)
}

func (m *inventoryRepositoryMock) GetByProductID(ctx context.Context, productID uuid.UUID) (*models.Inventory, error) {
	args := m.Called(ctx, productID)
	inventory, _ := args.Get(0).(*models.Inventory)
	return inventory, args.Error(1)
}

func (m *inventoryRepositoryMock) GetAll(ctx context.Context) ([]models.Inventory, error) {
	args := m.Called(ctx)
	inventories, _ := args.Get(0).([]models.Inventory)
	return inventories, args.Error(1)
}

func (m *inventoryRepositoryMock) ExistsByProductID(ctx context.Context, productID uuid.UUID) (bool, error) {
	args := m.Called(ctx, productID)
	return args.Bool(0), args.Error(1)
}

func (m *inventoryRepositoryMock) Update(ctx context.Context, inv *models.Inventory) error {
	return m.Called(ctx, inv).Error(0)
}

func (m *inventoryRepositoryMock) AddStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	return m.Called(ctx, productID, quantity).Error(0)
}

func (m *inventoryRepositoryMock) RemoveStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	return m.Called(ctx, productID, quantity).Error(0)
}

func (m *inventoryRepositoryMock) ReserveStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	return m.Called(ctx, productID, quantity).Error(0)
}

func (m *inventoryRepositoryMock) ReleaseReservedStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	return m.Called(ctx, productID, quantity).Error(0)
}

func (m *inventoryRepositoryMock) ConfirmReservedStock(ctx context.Context, productID uuid.UUID, quantity int) error {
	return m.Called(ctx, productID, quantity).Error(0)
}

type inventoryProductRepositoryMock struct {
	mock.Mock
}

func (m *inventoryProductRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	args := m.Called(ctx, id)
	product, _ := args.Get(0).(*models.Product)
	return product, args.Error(1)
}

func testInventory(productID uuid.UUID) *models.Inventory {
	return &models.Inventory{
		ProductID:         productID,
		AvailableQuantity: 10,
		ReservedQuantity:  3,
		Product: models.Product{
			BaseModel: models.BaseModel{ID: productID},
			Name:      "Laptop",
			SKU:       "SKU-1",
		},
	}
}

func TestInventoryServiceCreateInventory(t *testing.T) {
	ctx := context.Background()
	inventoryRepo := new(inventoryRepositoryMock)
	productRepo := new(inventoryProductRepositoryMock)
	service := NewService(inventoryRepo, productRepo, zap.NewNop())
	productID := uuid.New()

	productRepo.On("GetByID", ctx, productID).Return(&models.Product{BaseModel: models.BaseModel{ID: productID}}, nil).Once()
	inventoryRepo.On("ExistsByProductID", ctx, productID).Return(false, nil).Once()
	inventoryRepo.On("Create", ctx, mock.AnythingOfType("*models.Inventory")).Return(nil).Once()
	inventoryRepo.On("GetByProductID", ctx, productID).Return(testInventory(productID), nil).Once()

	response, err := service.CreateInventory(ctx, CreateInventoryRequest{
		ProductID:         productID,
		AvailableQuantity: 10,
	})

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 10, response.AvailableQuantity)
}

func TestInventoryServiceCreateInventoryMapsMissingProduct(t *testing.T) {
	ctx := context.Background()
	inventoryRepo := new(inventoryRepositoryMock)
	productRepo := new(inventoryProductRepositoryMock)
	service := NewService(inventoryRepo, productRepo, zap.NewNop())
	productID := uuid.New()

	productRepo.On("GetByID", ctx, productID).Return(nil, gorm.ErrRecordNotFound).Once()

	response, err := service.CreateInventory(ctx, CreateInventoryRequest{ProductID: productID})

	require.Error(t, err)
	assert.Nil(t, response)
	assert.ErrorIs(t, err, ErrProductNotFound)
}

func TestInventoryServiceRemoveStockRejectsInsufficientQuantity(t *testing.T) {
	ctx := context.Background()
	inventoryRepo := new(inventoryRepositoryMock)
	productRepo := new(inventoryProductRepositoryMock)
	service := NewService(inventoryRepo, productRepo, zap.NewNop())
	productID := uuid.New()

	inventoryRepo.On("GetByProductID", ctx, productID).Return(testInventory(productID), nil).Once()

	response, err := service.RemoveStock(ctx, productID, 20)

	require.Error(t, err)
	assert.Nil(t, response)
	assert.ErrorIs(t, err, ErrInsufficientStock)
}

func TestInventoryServiceReleaseReservedStockRejectsInvalidQuantity(t *testing.T) {
	service := NewService(new(inventoryRepositoryMock), new(inventoryProductRepositoryMock), zap.NewNop())

	response, err := service.ReleaseReservedStock(context.Background(), uuid.New(), 0)

	require.Error(t, err)
	assert.Nil(t, response)
	assert.ErrorIs(t, err, ErrInvalidQuantity)
}
