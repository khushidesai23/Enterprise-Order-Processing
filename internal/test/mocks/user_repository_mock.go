package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

type UserRepositoryMock struct {
	mock.Mock
}

func (m *UserRepositoryMock) Create(
	ctx context.Context,
	user *models.User,
) error {
	args := m.Called(ctx, user)

	return args.Error(0)
}

func (m *UserRepositoryMock) Update(
	ctx context.Context,
	user *models.User,
) error {
	args := m.Called(ctx, user)

	return args.Error(0)
}

func (m *UserRepositoryMock) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	args := m.Called(ctx, id)

	return args.Error(0)
}

func (m *UserRepositoryMock) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.User, error) {
	args := m.Called(ctx, id)

	user, _ := args.Get(0).(*models.User)

	return user, args.Error(1)
}

func (m *UserRepositoryMock) GetByIDTx(
	ctx context.Context,
	tx *gorm.DB,
	id uuid.UUID,
) (*models.User, error) {
	args := m.Called(ctx, tx, id)

	user, _ := args.Get(0).(*models.User)

	return user, args.Error(1)
}

func (m *UserRepositoryMock) GetByEmail(
	ctx context.Context,
	email string,
) (*models.User, error) {
	args := m.Called(ctx, email)

	user, _ := args.Get(0).(*models.User)

	return user, args.Error(1)
}

func (m *UserRepositoryMock) List(
	ctx context.Context,
) ([]models.User, error) {
	args := m.Called(ctx)

	users, _ := args.Get(0).([]models.User)

	return users, args.Error(1)
}
