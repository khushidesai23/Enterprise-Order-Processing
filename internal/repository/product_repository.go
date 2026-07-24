package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

// Create creates a new product.
func (r *ProductRepository) Create(product *models.Product) error {
	return r.db.Create(product).Error
}

// GetByID returns a product by its ID.
func (r *ProductRepository) GetByID(id uuid.UUID) (*models.Product, error) {
	var product models.Product

	err := r.db.
		Preload("Category").
		Preload("Inventory").
		First(&product, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}

	return &product, nil
}

// GetByIDTx returns a product by its ID inside a transaction.
func (r *ProductRepository) GetByIDTx(
	tx *gorm.DB,
	id uuid.UUID,
) (*models.Product, error) {
	var product models.Product

	err := tx.
		Preload("Category").
		Preload("Inventory").
		First(&product, "id = ?", id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}

	return &product, nil
}

// GetBySKU returns a product by SKU.
func (r *ProductRepository) GetBySKU(sku string) (*models.Product, error) {
	var product models.Product

	err := r.db.
		Preload("Category").
		Preload("Inventory").
		Where("sku = ?", sku).
		First(&product).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}

	return &product, nil
}

// GetAll returns all products.
func (r *ProductRepository) GetAll() ([]models.Product, error) {
	var products []models.Product

	err := r.db.
		Preload("Category").
		Preload("Inventory").
		Order("created_at DESC").
		Find(&products).Error

	if err != nil {
		return nil, err
	}

	return products, nil
}

// GetByCategory returns all products belonging to a category.
func (r *ProductRepository) GetByCategory(categoryID uuid.UUID) ([]models.Product, error) {
	var products []models.Product

	err := r.db.
		Preload("Category").
		Preload("Inventory").
		Where("category_id = ?", categoryID).
		Order("created_at DESC").
		Find(&products).Error

	if err != nil {
		return nil, err
	}

	return products, nil
}

// ExistsBySKU checks whether a SKU already exists.
func (r *ProductRepository) ExistsBySKU(sku string) (bool, error) {
	var count int64

	err := r.db.
		Model(&models.Product{}).
		Where("sku = ?", sku).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsBySKUExceptID checks duplicate SKU excluding current product.
func (r *ProductRepository) ExistsBySKUExceptID(id uuid.UUID, sku string) (bool, error) {
	var count int64

	err := r.db.
		Model(&models.Product{}).
		Where("sku = ? AND id <> ?", sku, id).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CategoryExists checks whether a category exists.
func (r *ProductRepository) CategoryExists(categoryID uuid.UUID) (bool, error) {
	var count int64

	err := r.db.
		Model(&models.Category{}).
		Where("id = ?", categoryID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Update updates an existing product.
func (r *ProductRepository) Update(product *models.Product) error {

	return r.db.Model(&models.Product{}).
		Where("id = ?", product.ID).
		Updates(map[string]interface{}{
			"name":        product.Name,
			"description": product.Description,
			"sku":         product.SKU,
			"price":       product.Price,
			"category_id": product.CategoryID,
		}).Error
}

// Delete soft deletes a product.
func (r *ProductRepository) Delete(product *models.Product) error {
	return r.db.Delete(product).Error
}
