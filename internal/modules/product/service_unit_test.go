package product

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

type productRepositoryMock struct {
	mock.Mock
}

func (m *productRepositoryMock) Create(ctx context.Context, product *models.Product) error {
	return m.Called(ctx, product).Error(0)
}

func (m *productRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error) {
	args := m.Called(ctx, id)
	product, _ := args.Get(0).(*models.Product)
	return product, args.Error(1)
}

func (m *productRepositoryMock) GetAll(ctx context.Context) ([]models.Product, error) {
	args := m.Called(ctx)
	products, _ := args.Get(0).([]models.Product)
	return products, args.Error(1)
}

func (m *productRepositoryMock) GetByCategory(ctx context.Context, categoryID uuid.UUID) ([]models.Product, error) {
	args := m.Called(ctx, categoryID)
	products, _ := args.Get(0).([]models.Product)
	return products, args.Error(1)
}

func (m *productRepositoryMock) ExistsBySKU(ctx context.Context, sku string) (bool, error) {
	args := m.Called(ctx, sku)
	return args.Bool(0), args.Error(1)
}

func (m *productRepositoryMock) ExistsBySKUExceptID(ctx context.Context, id uuid.UUID, sku string) (bool, error) {
	args := m.Called(ctx, id, sku)
	return args.Bool(0), args.Error(1)
}

func (m *productRepositoryMock) CategoryExists(ctx context.Context, categoryID uuid.UUID) (bool, error) {
	args := m.Called(ctx, categoryID)
	return args.Bool(0), args.Error(1)
}

func (m *productRepositoryMock) Update(ctx context.Context, product *models.Product) error {
	return m.Called(ctx, product).Error(0)
}

func (m *productRepositoryMock) Delete(ctx context.Context, product *models.Product) error {
	return m.Called(ctx, product).Error(0)
}

func testProduct(categoryID uuid.UUID) *models.Product {
	return &models.Product{
		BaseModel:  models.BaseModel{ID: uuid.New()},
		Name:       "Laptop",
		SKU:        "SKU-1",
		Price:      1299,
		CategoryID: categoryID,
		Category: models.Category{
			BaseModel: models.BaseModel{ID: categoryID},
			Name:      "Electronics",
		},
		Inventory: &models.Inventory{
			ProductID:         uuid.New(),
			AvailableQuantity: 5,
			ReservedQuantity:  1,
		},
	}
}

func TestProductServiceCreateProduct(t *testing.T) {
	ctx := context.Background()
	repo := new(productRepositoryMock)
	service := NewService(repo, zap.NewNop())
	categoryID := uuid.New()
	req := CreateProductRequest{
		Name:        "Laptop",
		Description: "Work machine",
		SKU:         "SKU-1",
		Price:       1299,
		CategoryID:  categoryID,
	}
	product := testProduct(categoryID)

	repo.On("CategoryExists", ctx, categoryID).Return(true, nil).Once()
	repo.On("ExistsBySKU", ctx, req.SKU).Return(false, nil).Once()
	repo.On("Create", ctx, mock.AnythingOfType("*models.Product")).Return(nil).Once()
	repo.On("GetByID", ctx, mock.AnythingOfType("uuid.UUID")).Return(product, nil).Once()

	response, err := service.CreateProduct(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, req.SKU, response.SKU)
	assert.Equal(t, req.Price, response.Price)
}

func TestProductServiceUpdateProductRejectsDuplicateSKU(t *testing.T) {
	ctx := context.Background()
	repo := new(productRepositoryMock)
	service := NewService(repo, zap.NewNop())
	categoryID := uuid.New()
	product := testProduct(categoryID)

	repo.On("GetByID", ctx, product.ID).Return(product, nil).Once()
	repo.On("CategoryExists", ctx, categoryID).Return(true, nil).Once()
	repo.On("ExistsBySKUExceptID", ctx, product.ID, "DUPLICATE").Return(true, nil).Once()

	response, err := service.UpdateProduct(ctx, product.ID, UpdateProductRequest{
		Name:        "Laptop",
		Description: "Updated",
		SKU:         "DUPLICATE",
		Price:       1499,
		CategoryID:  categoryID,
	})

	require.Error(t, err)
	assert.Nil(t, response)
	assert.ErrorIs(t, err, ErrProductSKUExists)
}

func TestProductServiceGetProductMapsNotFound(t *testing.T) {
	ctx := context.Background()
	repo := new(productRepositoryMock)
	service := NewService(repo, zap.NewNop())
	productID := uuid.New()

	repo.On("GetByID", ctx, productID).Return(nil, gorm.ErrRecordNotFound).Once()

	response, err := service.GetProduct(ctx, productID)

	require.Error(t, err)
	assert.Nil(t, response)
	assert.ErrorIs(t, err, ErrProductNotFound)
}

func TestProductServiceDeleteProductSuccess(t *testing.T) {
	ctx := context.Background()
	repo := new(productRepositoryMock)
	service := NewService(repo, zap.NewNop())
	product := testProduct(uuid.New())

	repo.On("GetByID", ctx, product.ID).Return(product, nil).Once()
	repo.On("Delete", ctx, product).Return(nil).Once()

	err := service.DeleteProduct(ctx, product.ID)

	require.NoError(t, err)
	repo.AssertExpectations(t)
}
