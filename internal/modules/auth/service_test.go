package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/khushidesai23/Enterprise-Order-Processing/config"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/test/mocks"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/utils"
)

func testAuthService(repo *mocks.UserRepositoryMock) *Service {
	cfg := &config.Config{
		JWTExpiration: time.Hour,
	}

	return NewService(
		repo,
		NewJWTManager("secret", time.Hour),
		cfg,
		zap.NewNop(),
	)
}

func testAuthUser(t *testing.T, active bool) *models.User {
	t.Helper()

	hash, err := utils.HashPassword("password123")
	require.NoError(t, err)

	return &models.User{
		BaseModel: models.BaseModel{
			ID: uuid.New(),
		},
		FirstName: "Khushi",
		LastName:  "Desai",
		Email:     "khushi@example.com",
		Password:  hash,
		IsActive:  active,
	}
}

func TestLogin(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(mocks.UserRepositoryMock)
		service := testAuthService(repo)
		user := testAuthUser(t, true)

		repo.On("GetByEmail", ctx, user.Email).Return(user, nil).Once()

		response, err := service.Login(ctx, LoginRequest{
			Email:    user.Email,
			Password: "password123",
		})

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "Bearer", response.TokenType)
		assert.Equal(t, user.Email, response.User.Email)
		assert.NotEmpty(t, response.Token)
		repo.AssertExpectations(t)
	})

	t.Run("repository error becomes invalid credentials", func(t *testing.T) {
		repo := new(mocks.UserRepositoryMock)
		service := testAuthService(repo)

		repo.On("GetByEmail", ctx, "missing@example.com").Return(nil, errors.New("db down")).Once()

		response, err := service.Login(ctx, LoginRequest{
			Email:    "missing@example.com",
			Password: "password123",
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		repo.AssertExpectations(t)
	})

	t.Run("inactive user", func(t *testing.T) {
		repo := new(mocks.UserRepositoryMock)
		service := testAuthService(repo)
		user := testAuthUser(t, false)

		repo.On("GetByEmail", ctx, user.Email).Return(user, nil).Once()

		response, err := service.Login(ctx, LoginRequest{
			Email:    user.Email,
			Password: "password123",
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.ErrorIs(t, err, ErrInactiveUser)
		repo.AssertExpectations(t)
	})

	t.Run("invalid password", func(t *testing.T) {
		repo := new(mocks.UserRepositoryMock)
		service := testAuthService(repo)
		user := testAuthUser(t, true)

		repo.On("GetByEmail", ctx, user.Email).Return(user, nil).Once()

		response, err := service.Login(ctx, LoginRequest{
			Email:    user.Email,
			Password: "wrong-password",
		})

		require.Error(t, err)
		assert.Nil(t, response)
		assert.ErrorIs(t, err, ErrInvalidCredentials)
		repo.AssertExpectations(t)
	})
}

func TestGetCurrentUser(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		repo := new(mocks.UserRepositoryMock)
		service := testAuthService(repo)
		user := testAuthUser(t, true)

		repo.On("GetByID", ctx, user.ID).Return(user, nil).Once()

		response, err := service.GetCurrentUser(ctx, user.ID.String())

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, user.Email, response.Email)
		repo.AssertExpectations(t)
	})

	t.Run("invalid uuid", func(t *testing.T) {
		repo := new(mocks.UserRepositoryMock)
		service := testAuthService(repo)

		response, err := service.GetCurrentUser(ctx, "bad-uuid")

		require.Error(t, err)
		assert.Nil(t, response)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
}
