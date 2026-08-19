package mocks

import (
	"context"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

type CategoryRepositoryMock struct {
	mock.Mock
}

func (m *CategoryRepositoryMock) Create(
	ctx context.Context,
	category *models.Category,
) error {
	args := m.Called(ctx, category)

	return args.Error(0)
}

func (m *CategoryRepositoryMock) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.Category, error) {
	args := m.Called(ctx, id)

	category, _ := args.Get(0).(*models.Category)

	return category, args.Error(1)
}

func (m *CategoryRepositoryMock) GetAll(
	ctx context.Context,
) ([]models.Category, error) {
	args := m.Called(ctx)

	categories, _ := args.Get(0).([]models.Category)

	return categories, args.Error(1)
}

func (m *CategoryRepositoryMock) ExistsByName(
	ctx context.Context,
	name string,
) (bool, error) {
	args := m.Called(ctx, name)

	return args.Bool(0), args.Error(1)
}

func (m *CategoryRepositoryMock) ExistsByNameExceptID(
	ctx context.Context,
	id uuid.UUID,
	name string,
) (bool, error) {
	args := m.Called(ctx, id, name)

	return args.Bool(0), args.Error(1)
}

func (m *CategoryRepositoryMock) HasProducts(
	ctx context.Context,
	id uuid.UUID,
) (bool, error) {
	args := m.Called(ctx, id)

	return args.Bool(0), args.Error(1)
}

func (m *CategoryRepositoryMock) Update(
	ctx context.Context,
	category *models.Category,
) error {
	args := m.Called(ctx, category)

	return args.Error(0)
}

func (m *CategoryRepositoryMock) Delete(
	ctx context.Context,
	category *models.Category,
) error {
	args := m.Called(ctx, category)

	return args.Error(0)
}