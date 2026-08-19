package category

import (
	"context"

	"github.com/google/uuid"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

type CategoryRepository interface {
	Create(
		ctx context.Context,
		category *models.Category,
	) error

	GetByID(
		ctx context.Context,
		id uuid.UUID,
	) (*models.Category, error)

	GetAll(
		ctx context.Context,
	) ([]models.Category, error)

	ExistsByName(
		ctx context.Context,
		name string,
	) (bool, error)

	ExistsByNameExceptID(
		ctx context.Context,
		id uuid.UUID,
		name string,
	) (bool, error)

	HasProducts(
		ctx context.Context,
		id uuid.UUID,
	) (bool, error)

	Update(
		ctx context.Context,
		category *models.Category,
	) error

	Delete(
		ctx context.Context,
		category *models.Category,
	) error
}