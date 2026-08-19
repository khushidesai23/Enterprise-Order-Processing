package category

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

type categoryRepositoryMock struct {
	mock.Mock
}

func (m *categoryRepositoryMock) Create(ctx context.Context, category *models.Category) error {
	return m.Called(ctx, category).Error(0)
}

func (m *categoryRepositoryMock) GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error) {
	args := m.Called(ctx, id)
	category, _ := args.Get(0).(*models.Category)
	return category, args.Error(1)
}

func (m *categoryRepositoryMock) GetAll(ctx context.Context) ([]models.Category, error) {
	args := m.Called(ctx)
	categories, _ := args.Get(0).([]models.Category)
	return categories, args.Error(1)
}

func (m *categoryRepositoryMock) ExistsByName(ctx context.Context, name string) (bool, error) {
	args := m.Called(ctx, name)
	return args.Bool(0), args.Error(1)
}

func (m *categoryRepositoryMock) ExistsByNameExceptID(ctx context.Context, id uuid.UUID, name string) (bool, error) {
	args := m.Called(ctx, id, name)
	return args.Bool(0), args.Error(1)
}

func (m *categoryRepositoryMock) HasProducts(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *categoryRepositoryMock) Update(ctx context.Context, category *models.Category) error {
	return m.Called(ctx, category).Error(0)
}

func (m *categoryRepositoryMock) Delete(ctx context.Context, category *models.Category) error {
	return m.Called(ctx, category).Error(0)
}

func TestCategoryServiceCreateCategory(t *testing.T) {
	ctx := context.Background()
	repo := new(categoryRepositoryMock)
	service := NewService(repo, zap.NewNop())

	req := CreateCategoryRequest{Name: "Electronics", Description: "Devices"}

	repo.On("ExistsByName", ctx, req.Name).Return(false, nil).Once()
	repo.On("Create", ctx, mock.AnythingOfType("*models.Category")).Return(nil).Once()
	repo.On("GetByID", ctx, mock.AnythingOfType("uuid.UUID")).Return(&models.Category{
		BaseModel:   models.BaseModel{ID: uuid.New()},
		Name:        req.Name,
		Description: req.Description,
	}, nil).Once()

	response, err := service.CreateCategory(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, req.Name, response.Name)
	repo.AssertExpectations(t)
}

func TestCategoryServiceCreateCategoryRejectsDuplicateName(t *testing.T) {
	ctx := context.Background()
	repo := new(categoryRepositoryMock)
	service := NewService(repo, zap.NewNop())

	repo.On("ExistsByName", ctx, "Electronics").Return(true, nil).Once()

	response, err := service.CreateCategory(ctx, CreateCategoryRequest{Name: "Electronics"})

	require.Error(t, err)
	assert.Nil(t, response)
	assert.ErrorIs(t, err, ErrCategoryNameExists)
}

func TestCategoryServiceGetCategoryMapsNotFound(t *testing.T) {
	ctx := context.Background()
	repo := new(categoryRepositoryMock)
	service := NewService(repo, zap.NewNop())
	categoryID := uuid.New()

	repo.On("GetByID", ctx, categoryID).Return(nil, gorm.ErrRecordNotFound).Once()

	response, err := service.GetCategory(ctx, categoryID)

	require.Error(t, err)
	assert.Nil(t, response)
	assert.ErrorIs(t, err, ErrCategoryNotFound)
}

func TestCategoryServiceDeleteCategoryBlocksCategoryInUse(t *testing.T) {
	ctx := context.Background()
	repo := new(categoryRepositoryMock)
	service := NewService(repo, zap.NewNop())
	category := &models.Category{BaseModel: models.BaseModel{ID: uuid.New()}, Name: "Electronics"}

	repo.On("GetByID", ctx, category.ID).Return(category, nil).Once()
	repo.On("HasProducts", ctx, category.ID).Return(true, nil).Once()

	err := service.DeleteCategory(ctx, category.ID)

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrCategoryInUse)
}

func TestCategoryServiceUpdateCategoryPropagatesRepositoryError(t *testing.T) {
	ctx := context.Background()
	repo := new(categoryRepositoryMock)
	service := NewService(repo, zap.NewNop())
	category := &models.Category{BaseModel: models.BaseModel{ID: uuid.New()}, Name: "Old"}
	expectedErr := errors.New("write failed")

	repo.On("GetByID", ctx, category.ID).Return(category, nil).Once()
	repo.On("ExistsByNameExceptID", ctx, category.ID, "New").Return(false, nil).Once()
	repo.On("Update", ctx, mock.AnythingOfType("*models.Category")).Return(expectedErr).Once()

	response, err := service.UpdateCategory(ctx, category.ID, UpdateCategoryRequest{Name: "New"})

	require.Error(t, err)
	assert.Nil(t, response)
	assert.ErrorIs(t, err, expectedErr)
}
