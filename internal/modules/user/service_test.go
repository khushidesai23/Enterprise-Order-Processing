package user

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

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/test/mocks"
)

func newTestService(
	repo *mocks.UserRepositoryMock,
) Service {
	return NewService(
		repo,
		zap.NewNop(),
	)
}

func testUser() *models.User {
	return &models.User{
		BaseModel: models.BaseModel{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		FirstName: "Khushi",
		LastName:  "Desai",
		Email:     "khushi@example.com",
		Password:  "hashed-password",
		IsActive:  true,
	}
}

func TestCreateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(mocks.UserRepositoryMock)
		service := newTestService(repo)

		req := CreateUserRequest{
			FirstName: "Khushi",
			LastName:  "Desai",
			Email:     "khushi@example.com",
			Password:  "password123",
		}

		repo.
			On("GetByEmail", ctx, "khushi@example.com").
			Return(nil, nil).
			Once()

		repo.
			On("Create", ctx, mock.MatchedBy(func(user *models.User) bool {
				return user.FirstName == "Khushi" &&
					user.LastName == "Desai" &&
					user.Email == "khushi@example.com" &&
					user.IsActive
			})).
			Return(nil).
			Once()

		response, err := service.Create(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)

		assert.Equal(t, "Khushi", response.FirstName)
		assert.Equal(t, "Desai", response.LastName)
		assert.Equal(t, "khushi@example.com", response.Email)
		assert.True(t, response.IsActive)

		repo.AssertExpectations(t)
	})

	t.Run("duplicate email", func(t *testing.T) {
		repo := new(mocks.UserRepositoryMock)
		service := newTestService(repo)

		req := CreateUserRequest{
			FirstName: "Khushi",
			Email:     "khushi@example.com",
			Password:  "password123",
		}

		repo.
			On("GetByEmail", ctx, req.Email).
			Return(testUser(), nil).
			Once()

		response, err := service.Create(ctx, req)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.ErrorIs(t, err, ErrUserAlreadyExists)

		repo.AssertExpectations(t)
	})

	t.Run("repository error", func(t *testing.T) {
		repo := new(mocks.UserRepositoryMock)
		service := newTestService(repo)

		req := CreateUserRequest{
			FirstName: "Khushi",
			Email:     "khushi@example.com",
			Password:  "password123",
		}

		expectedErr := errors.New("database error")

		repo.
			On("GetByEmail", ctx, req.Email).
			Return(nil, expectedErr).
			Once()

		response, err := service.Create(ctx, req)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.ErrorIs(t, err, expectedErr)

		repo.AssertExpectations(t)
	})
}

func TestGetUserByID(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(mocks.UserRepositoryMock)
		service := newTestService(repo)

		expectedUser := testUser()

		repo.
			On("GetByID", ctx, expectedUser.ID).
			Return(expectedUser, nil).
			Once()

		response, err := service.GetByID(ctx, expectedUser.ID)

		require.NoError(t, err)
		require.NotNil(t, response)

		assert.Equal(t, expectedUser.ID.String(), response.ID)
		assert.Equal(t, expectedUser.Email, response.Email)

		repo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		repo := new(mocks.UserRepositoryMock)
		service := newTestService(repo)

		userID := uuid.New()

		repo.
			On("GetByID", ctx, userID).
			Return(nil, nil).
			Once()

		response, err := service.GetByID(ctx, userID)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.ErrorIs(t, err, ErrUserNotFound)

		repo.AssertExpectations(t)
	})
}

func TestUpdateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(mocks.UserRepositoryMock)
		service := newTestService(repo)

		existingUser := testUser()

		isActive := false

		req := UpdateUserRequest{
			FirstName: "Updated",
			LastName:  "User",
			IsActive:  &isActive,
		}

		repo.
			On("GetByID", ctx, existingUser.ID).
			Return(existingUser, nil).
			Once()

		repo.
			On("Update", ctx, mock.AnythingOfType("*models.User")).
			Return(nil).
			Once()

		response, err := service.Update(ctx, existingUser.ID, req)

		require.NoError(t, err)
		require.NotNil(t, response)

		assert.Equal(t, "Updated", response.FirstName)
		assert.Equal(t, "User", response.LastName)
		assert.False(t, response.IsActive)

		repo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		repo := new(mocks.UserRepositoryMock)
		service := newTestService(repo)

		userID := uuid.New()

		repo.
			On("GetByID", ctx, userID).
			Return(nil, nil).
			Once()

		response, err := service.Update(
			ctx,
			userID,
			UpdateUserRequest{
				FirstName: "Updated",
			},
		)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.ErrorIs(t, err, ErrUserNotFound)

		repo.AssertExpectations(t)
	})
}

func TestDeleteUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(mocks.UserRepositoryMock)
		service := newTestService(repo)

		existingUser := testUser()

		repo.
			On("GetByID", ctx, existingUser.ID).
			Return(existingUser, nil).
			Once()

		repo.
			On("Delete", ctx, existingUser.ID).
			Return(nil).
			Once()

		err := service.Delete(ctx, existingUser.ID)

		require.NoError(t, err)

		repo.AssertExpectations(t)
	})

	t.Run("user not found", func(t *testing.T) {
		repo := new(mocks.UserRepositoryMock)
		service := newTestService(repo)

		userID := uuid.New()

		repo.
			On("GetByID", ctx, userID).
			Return(nil, nil).
			Once()

		err := service.Delete(ctx, userID)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrUserNotFound)

		repo.AssertExpectations(t)
	})
}
