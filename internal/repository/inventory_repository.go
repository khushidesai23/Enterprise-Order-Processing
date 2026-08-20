package repository

import (
	"context"
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
func (r *InventoryRepository) Create(
	ctx context.Context,
	inventory *models.Inventory,
) error {

	return r.db.WithContext(ctx).
		Create(inventory).
		Error
}

// GetByProductID returns inventory by product id.
func (r *InventoryRepository) GetByProductID(
	ctx context.Context,
	productID uuid.UUID,
) (*models.Inventory, error) {

	var inventory models.Inventory

	err := r.db.WithContext(ctx).
		Preload("Product").
		First(
			&inventory,
			"product_id = ?",
			productID,
		).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		return nil, err
	}

	return &inventory, nil
}

// GetByProductIDTx returns inventory within a transaction.
func (r *InventoryRepository) GetByProductIDTx(
	ctx context.Context,
	tx *gorm.DB,
	productID uuid.UUID,
) (*models.Inventory, error) {

	var inventory models.Inventory

	err := tx.WithContext(ctx).
		Preload("Product").
		First(
			&inventory,
			"product_id = ?",
			productID,
		).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		return nil, err
	}

	return &inventory, nil
}

// GetAll returns all inventories.
func (r *InventoryRepository) GetAll(
	ctx context.Context,
) ([]models.Inventory, error) {

	var inventories []models.Inventory

	err := r.db.WithContext(ctx).
		Preload("Product").
		Order("created_at DESC").
		Find(&inventories).
		Error

	if err != nil {
		return nil, err
	}

	return inventories, nil
}

// ExistsByProductID checks whether inventory already exists.
func (r *InventoryRepository) ExistsByProductID(
	ctx context.Context,
	productID uuid.UUID,
) (bool, error) {

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		Count(&count).
		Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ProductExists checks whether a product exists.
func (r *InventoryRepository) ProductExists(
	ctx context.Context,
	productID uuid.UUID,
) (bool, error) {

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where("id = ?", productID).
		Count(&count).
		Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Update updates inventory.
func (r *InventoryRepository) Update(
	ctx context.Context,
	inv *models.Inventory,
) error {

	return r.db.WithContext(ctx).
		Model(&models.Inventory{}).
		Where("product_id = ?", inv.ProductID).
		Updates(map[string]interface{}{
			"available_quantity": inv.AvailableQuantity,
			"reserved_quantity":  inv.ReservedQuantity,
		}).
		Error
}

// Delete deletes inventory.
func (r *InventoryRepository) Delete(
	ctx context.Context,
	inventory *models.Inventory,
) error {

	return r.db.WithContext(ctx).
		Delete(inventory).
		Error
}

//
// Business Operations
//

// AddStock increases available stock.
func (r *InventoryRepository) AddStock(
	ctx context.Context,
	productID uuid.UUID,
	quantity int,
) error {

	return r.db.WithContext(ctx).
		Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		UpdateColumn(
			"available_quantity",
			gorm.Expr(
				"available_quantity + ?",
				quantity,
			),
		).
		Error
}

// AddStockTx increases available stock inside a transaction.
func (r *InventoryRepository) AddStockTx(
	ctx context.Context,
	tx *gorm.DB,
	productID uuid.UUID,
	quantity int,
) error {

	return tx.WithContext(ctx).
		Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		UpdateColumn(
			"available_quantity",
			gorm.Expr(
				"available_quantity + ?",
				quantity,
			),
		).
		Error
}

// RemoveStock decreases available stock.
func (r *InventoryRepository) RemoveStock(
	ctx context.Context,
	productID uuid.UUID,
	quantity int,
) error {

	return r.db.WithContext(ctx).
		Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		UpdateColumn(
			"available_quantity",
			gorm.Expr(
				"available_quantity - ?",
				quantity,
			),
		).
		Error
}

// RemoveStockTx decreases available stock inside a transaction.
func (r *InventoryRepository) RemoveStockTx(
	ctx context.Context,
	tx *gorm.DB,
	productID uuid.UUID,
	quantity int,
) error {

	return tx.WithContext(ctx).
		Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		UpdateColumn(
			"available_quantity",
			gorm.Expr(
				"available_quantity - ?",
				quantity,
			),
		).
		Error
}

// ReserveStock moves stock from available -> reserved.
func (r *InventoryRepository) ReserveStock(
	ctx context.Context,
	productID uuid.UUID,
	quantity int,
) error {

	return r.db.WithContext(ctx).
		Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		Updates(map[string]interface{}{
			"available_quantity": gorm.Expr(
				"available_quantity - ?",
				quantity,
			),
			"reserved_quantity": gorm.Expr(
				"reserved_quantity + ?",
				quantity,
			),
		}).
		Error
}

// ReserveStockTx moves stock from available -> reserved inside a transaction.
func (r *InventoryRepository) ReserveStockTx(
	ctx context.Context,
	tx *gorm.DB,
	productID uuid.UUID,
	quantity int,
) error {

	return tx.WithContext(ctx).
		Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		Updates(map[string]interface{}{
			"available_quantity": gorm.Expr(
				"available_quantity - ?",
				quantity,
			),
			"reserved_quantity": gorm.Expr(
				"reserved_quantity + ?",
				quantity,
			),
		}).
		Error
}

// ReleaseReservedStock moves stock from reserved -> available.
func (r *InventoryRepository) ReleaseReservedStock(
	ctx context.Context,
	productID uuid.UUID,
	quantity int,
) error {

	return r.db.WithContext(ctx).
		Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		Updates(map[string]interface{}{
			"available_quantity": gorm.Expr(
				"available_quantity + ?",
				quantity,
			),
			"reserved_quantity": gorm.Expr(
				"reserved_quantity - ?",
				quantity,
			),
		}).
		Error
}

// ReleaseReservedStockTx moves stock from reserved -> available inside a transaction.
func (r *InventoryRepository) ReleaseReservedStockTx(
	ctx context.Context,
	tx *gorm.DB,
	productID uuid.UUID,
	quantity int,
) error {

	return tx.WithContext(ctx).
		Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		Updates(map[string]interface{}{
			"available_quantity": gorm.Expr(
				"available_quantity + ?",
				quantity,
			),
			"reserved_quantity": gorm.Expr(
				"reserved_quantity - ?",
				quantity,
			),
		}).
		Error
}

// ConfirmReservedStock deducts reserved stock permanently.
func (r *InventoryRepository) ConfirmReservedStock(
	ctx context.Context,
	productID uuid.UUID,
	quantity int,
) error {

	return r.db.WithContext(ctx).
		Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		UpdateColumn(
			"reserved_quantity",
			gorm.Expr(
				"reserved_quantity - ?",
				quantity,
			),
		).
		Error
}

// ConfirmReservedStockTx deducts reserved stock permanently inside a transaction.
func (r *InventoryRepository) ConfirmReservedStockTx(
	ctx context.Context,
	tx *gorm.DB,
	productID uuid.UUID,
	quantity int,
) error {

	return tx.WithContext(ctx).
		Model(&models.Inventory{}).
		Where("product_id = ?", productID).
		UpdateColumn(
			"reserved_quantity",
			gorm.Expr(
				"reserved_quantity - ?",
				quantity,
			),
		).
		Error
}
