package repository

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

type InventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) *InventoryRepository {
	return &InventoryRepository{
		db: db,
	}
}

// Create creates inventory for a product.
func (r *InventoryRepository) Create(inventory *models.Inventory) error {
	return r.db.Create(inventory).Error
}

// GetByProductID returns inventory by product id.
func (r *InventoryRepository) GetByProductID(productID uuid.UUID) (*models.Inventory, error) {
	var inventory models.Inventory

	err := r.db.
		Preload("Product").
		First(&inventory, "product_id = ?", productID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}

	return &inventory, nil
}

// GetAll returns all inventories.
func (r *InventoryRepository) GetAll() ([]models.Inventory, error) {
	var inventories []models.Inventory

	err := r.db.
		Preload("Product").
		Order("created_at DESC").
		Find(&inventories).Error

	if err != nil {
		return nil, err
	}

	return inventories, nil
}

// ExistsByProductID checks whether inventory already exists.
func (r *InventoryRepository) ExistsByProductID(productID uuid.UUID) (bool, error) {
	var count int64

	err := r.db.
		Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ProductExists checks whether a product exists.
func (r *InventoryRepository) ProductExists(productID uuid.UUID) (bool, error) {
	var count int64

	err := r.db.
		Model(&models.Product{}).
		Where("id = ?", productID).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Update updates inventory.
func (r *InventoryRepository) Update(inv *models.Inventory) error {

    return r.db.Model(&models.Inventory{}).
        Where("product_id = ?", inv.ProductID).
        Updates(map[string]interface{}{
            "available_quantity": inv.AvailableQuantity,
            "reserved_quantity":  inv.ReservedQuantity,
        }).Error
}

// Delete deletes inventory.
func (r *InventoryRepository) Delete(inventory *models.Inventory) error {
	return r.db.Delete(inventory).Error
}

//
// Business Operations
//

// AddStock increases available stock.
func (r *InventoryRepository) AddStock(productID uuid.UUID, quantity int) error {
	return r.db.Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		UpdateColumn("available_quantity", gorm.Expr("available_quantity + ?", quantity)).
		Error
}

// RemoveStock decreases available stock.
func (r *InventoryRepository) RemoveStock(productID uuid.UUID, quantity int) error {
	return r.db.Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		UpdateColumn("available_quantity", gorm.Expr("available_quantity - ?", quantity)).
		Error
}

// ReserveStock moves stock from available -> reserved.
func (r *InventoryRepository) ReserveStock(productID uuid.UUID, quantity int) error {
	return r.db.Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		Updates(map[string]interface{}{
			"available_quantity": gorm.Expr("available_quantity - ?", quantity),
			"reserved_quantity":  gorm.Expr("reserved_quantity + ?", quantity),
		}).Error
}

// ReleaseReservedStock moves stock from reserved -> available.
func (r *InventoryRepository) ReleaseReservedStock(productID uuid.UUID, quantity int) error {
	return r.db.Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		Updates(map[string]interface{}{
			"available_quantity": gorm.Expr("available_quantity + ?", quantity),
			"reserved_quantity":  gorm.Expr("reserved_quantity - ?", quantity),
		}).Error
}

// ConfirmReservedStock deducts reserved stock permanently.
// This is called after successful payment.
func (r *InventoryRepository) ConfirmReservedStock(productID uuid.UUID, quantity int) error {
	return r.db.Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		UpdateColumn("reserved_quantity", gorm.Expr("reserved_quantity - ?", quantity)).
		Error
}