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
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/test/mocks"
)

func newTestService(repo *mocks.CategoryRepositoryMock) *Service {
	return NewService(
		repo,
		zap.NewNop(),
	)
}

func TestCreateCategory(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(mocks.CategoryRepositoryMock)
		service := newTestService(repo)

		req := CreateCategoryRequest{
			Name:        "Electronics",
			Description: "Electronic products",
		}

		repo.
			On("ExistsByName", ctx, req.Name).
			Return(false, nil).
			Once()

		repo.
			On("Create", ctx, mock.AnythingOfType("*models.Category")).
			Return(nil).
			Once()

		repo.
			On("GetByID", ctx, mock.Anything).
			Return(&models.Category{
				Name:        req.Name,
				Description: req.Description,
			}, nil).
			Once()

		response, err := service.CreateCategory(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)

		assert.Equal(t, req.Name, response.Name)
		assert.Equal(t, req.Description, response.Description)

		repo.AssertExpectations(t)
	})

	t.Run("duplicate category name", func(t *testing.T) {
		repo := new(mocks.CategoryRepositoryMock)
		service := newTestService(repo)

		req := CreateCategoryRequest{
			Name: "Electronics",
		}

		repo.
			On("ExistsByName", ctx, req.Name).
			Return(true, nil).
			Once()

		response, err := service.CreateCategory(ctx, req)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.ErrorIs(t, err, ErrCategoryNameExists)

		repo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := new(mocks.CategoryRepositoryMock)
		service := newTestService(repo)

		req := CreateCategoryRequest{
			Name: "Electronics",
		}

		expectedErr := errors.New("database error")

		repo.
			On("ExistsByName", ctx, req.Name).
			Return(false, expectedErr).
			Once()

		response, err := service.CreateCategory(ctx, req)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.ErrorIs(t, err, expectedErr)

		repo.AssertExpectations(t)
	})
}

func TestGetCategory(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(mocks.CategoryRepositoryMock)
		service := newTestService(repo)

		categoryID := uuid.New()

		expectedCategory := &models.Category{
			BaseModel: models.BaseModel{
				ID: categoryID,
			},
			Name:        "Electronics",
			Description: "Electronic products",
		}

		repo.
			On("GetByID", ctx, categoryID).
			Return(expectedCategory, nil).
			Once()

		response, err := service.GetCategory(ctx, categoryID)

		require.NoError(t, err)
		require.NotNil(t, response)

		assert.Equal(t, categoryID, response.ID)
		assert.Equal(t, expectedCategory.Name, response.Name)

		repo.AssertExpectations(t)
	})

	t.Run("category not found", func(t *testing.T) {
		repo := new(mocks.CategoryRepositoryMock)
		service := newTestService(repo)

		categoryID := uuid.New()

		repo.
			On("GetByID", ctx, categoryID).
			Return(nil, gorm.ErrRecordNotFound).
			Once()

		response, err := service.GetCategory(ctx, categoryID)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.ErrorIs(t, err, ErrCategoryNotFound)

		repo.AssertExpectations(t)
	})
}

func TestGetCategories(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(mocks.CategoryRepositoryMock)
		service := newTestService(repo)

		expectedCategories := []models.Category{
			{
				BaseModel: models.BaseModel{
					ID: uuid.New(),
				},
				Name: "Electronics",
			},
			{
				BaseModel: models.BaseModel{
					ID: uuid.New(),
				},
				Name: "Clothing",
			},
		}

		repo.
			On("GetAll", ctx).
			Return(expectedCategories, nil).
			Once()

		response, err := service.GetCategories(ctx)

		require.NoError(t, err)
		assert.Len(t, response, 2)

		repo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := new(mocks.CategoryRepositoryMock)
		service := newTestService(repo)

		expectedErr := errors.New("database error")

		repo.
			On("GetAll", ctx).
			Return(nil, expectedErr).
			Once()

		response, err := service.GetCategories(ctx)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.ErrorIs(t, err, expectedErr)

		repo.AssertExpectations(t)
	})
}

func TestUpdateCategory(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(mocks.CategoryRepositoryMock)
		service := newTestService(repo)

		categoryID := uuid.New()

		req := UpdateCategoryRequest{
			Name:        "Updated Electronics",
			Description: "Updated description",
		}

		existingCategory := &models.Category{
			BaseModel: models.BaseModel{
				ID: categoryID,
			},
			Name:        "Electronics",
			Description: "Old description",
		}

		repo.
			On("GetByID", ctx, categoryID).
			Return(existingCategory, nil).
			Twice()

		repo.
			On("ExistsByNameExceptID", ctx, categoryID, req.Name).
			Return(false, nil).
			Once()

		repo.
			On("Update", ctx, mock.AnythingOfType("*models.Category")).
			Return(nil).
			Once()

		response, err := service.UpdateCategory(
			ctx,
			categoryID,
			req,
		)

		require.NoError(t, err)
		require.NotNil(t, response)

		repo.AssertExpectations(t)
	})

	t.Run("duplicate category name", func(t *testing.T) {
		repo := new(mocks.CategoryRepositoryMock)
		service := newTestService(repo)

		categoryID := uuid.New()

		req := UpdateCategoryRequest{
			Name: "Existing Category",
		}

		repo.
			On("GetByID", ctx, categoryID).
			Return(&models.Category{
				BaseModel: models.BaseModel{
					ID: categoryID,
				},
			}, nil).
			Once()

		repo.
			On("ExistsByNameExceptID", ctx, categoryID, req.Name).
			Return(true, nil).
			Once()

		response, err := service.UpdateCategory(
			ctx,
			categoryID,
			req,
		)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.ErrorIs(t, err, ErrCategoryNameExists)

		repo.AssertExpectations(t)
	})
}

func TestDeleteCategory(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(mocks.CategoryRepositoryMock)
		service := newTestService(repo)

		categoryID := uuid.New()

		category := &models.Category{
			BaseModel: models.BaseModel{
				ID: categoryID,
			},
			Name: "Electronics",
		}

		repo.
			On("GetByID", ctx, categoryID).
			Return(category, nil).
			Once()

		repo.
			On("HasProducts", ctx, categoryID).
			Return(false, nil).
			Once()

		repo.
			On("Delete", ctx, category).
			Return(nil).
			Once()

		err := service.DeleteCategory(ctx, categoryID)

		require.NoError(t, err)

		repo.AssertExpectations(t)
	})

	t.Run("category has products", func(t *testing.T) {
		repo := new(mocks.CategoryRepositoryMock)
		service := newTestService(repo)

		categoryID := uuid.New()

		category := &models.Category{
			BaseModel: models.BaseModel{
				ID: categoryID,
			},
		}

		repo.
			On("GetByID", ctx, categoryID).
			Return(category, nil).
			Once()

		repo.
			On("HasProducts", ctx, categoryID).
			Return(true, nil).
			Once()

		err := service.DeleteCategory(ctx, categoryID)

		require.Error(t, err)

		repo.AssertExpectations(t)
	})
}
