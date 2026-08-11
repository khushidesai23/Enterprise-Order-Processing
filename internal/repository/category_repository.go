package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

type CategoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{
		db: db,
	}
}

// Create creates a new category.
func (r *CategoryRepository) Create(
	ctx context.Context,
	category *models.Category,
) error {

	return r.db.WithContext(ctx).
		Create(category).
		Error
}

// GetByID returns a category by ID.
func (r *CategoryRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.Category, error) {

	var category models.Category

	err := r.db.WithContext(ctx).
		Preload("Products").
		First(&category, "id = ?", id).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		return nil, err
	}

	return &category, nil
}

// GetByName returns a category by name.
func (r *CategoryRepository) GetByName(
	ctx context.Context,
	name string,
) (*models.Category, error) {

	var category models.Category

	err := r.db.WithContext(ctx).
		Preload("Products").
		Where("name = ?", name).
		First(&category).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		return nil, err
	}

	return &category, nil
}

// GetAll returns all categories.
func (r *CategoryRepository) GetAll(
	ctx context.Context,
) ([]models.Category, error) {

	var categories []models.Category

	err := r.db.WithContext(ctx).
		Preload("Products").
		Order("name ASC").
		Find(&categories).
		Error

	if err != nil {
		return nil, err
	}

	return categories, nil
}

// ExistsByName checks if a category name already exists.
func (r *CategoryRepository) ExistsByName(
	ctx context.Context,
	name string,
) (bool, error) {

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Category{}).
		Where("LOWER(name) = LOWER(?)", name).
		Count(&count).
		Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByNameExceptID checks whether another category already
// exists with the given name.
func (r *CategoryRepository) ExistsByNameExceptID(
	ctx context.Context,
	id uuid.UUID,
	name string,
) (bool, error) {

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Category{}).
		Where(
			"LOWER(name) = LOWER(?) AND id <> ?",
			name,
			id,
		).
		Count(&count).
		Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// HasProducts returns true if the category is associated
// with one or more products.
func (r *CategoryRepository) HasProducts(
	ctx context.Context,
	id uuid.UUID,
) (bool, error) {

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where("category_id = ?", id).
		Count(&count).
		Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Update updates a category.
func (r *CategoryRepository) Update(
	ctx context.Context,
	category *models.Category,
) error {

	return r.db.WithContext(ctx).
		Save(category).
		Error
}

// Delete soft deletes a category.
func (r *CategoryRepository) Delete(
	ctx context.Context,
	category *models.Category,
) error {

	return r.db.WithContext(ctx).
		Delete(category).
		Error
}