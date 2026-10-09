package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

// ErrInsufficientQuantity is returned when a stock decrement would make the
// guarded quantity negative (or the inventory row does not exist).
var ErrInsufficientQuantity = errors.New("insufficient inventory quantity")

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

// GetByProductIDTx returns inventory within a transaction and locks the row
// (FOR UPDATE). The product association is not loaded.
func (r *InventoryRepository) GetByProductIDTx(
	ctx context.Context,
	tx *gorm.DB,
	productID uuid.UUID,
) (*models.Inventory, error) {

	var inventory models.Inventory

	err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
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

// RemoveStock decreases available stock.
func (r *InventoryRepository) RemoveStock(
	ctx context.Context,
	productID uuid.UUID,
	quantity int,
) error {

	return r.RemoveStockTx(ctx, r.db, productID, quantity)
}

// RemoveStockTx decreases available stock inside a transaction.
func (r *InventoryRepository) RemoveStockTx(
	ctx context.Context,
	tx *gorm.DB,
	productID uuid.UUID,
	quantity int,
) error {

	return guardedUpdate(ctx, tx, productID, "available_quantity", quantity,
		map[string]interface{}{
			"available_quantity": gorm.Expr("available_quantity - ?", quantity),
		},
	)
}

// ReserveStock moves stock from available -> reserved.
func (r *InventoryRepository) ReserveStock(
	ctx context.Context,
	productID uuid.UUID,
	quantity int,
) error {

	return r.ReserveStockTx(ctx, r.db, productID, quantity)
}

// ReserveStockTx moves stock from available -> reserved inside a transaction.
func (r *InventoryRepository) ReserveStockTx(
	ctx context.Context,
	tx *gorm.DB,
	productID uuid.UUID,
	quantity int,
) error {

	return guardedUpdate(ctx, tx, productID, "available_quantity", quantity,
		map[string]interface{}{
			"available_quantity": gorm.Expr("available_quantity - ?", quantity),
			"reserved_quantity":  gorm.Expr("reserved_quantity + ?", quantity),
		},
	)
}

// ReleaseReservedStock moves stock from reserved -> available.
func (r *InventoryRepository) ReleaseReservedStock(
	ctx context.Context,
	productID uuid.UUID,
	quantity int,
) error {

	return r.ReleaseReservedStockTx(ctx, r.db, productID, quantity)
}

// ReleaseReservedStockTx moves stock from reserved -> available inside a transaction.
func (r *InventoryRepository) ReleaseReservedStockTx(
	ctx context.Context,
	tx *gorm.DB,
	productID uuid.UUID,
	quantity int,
) error {

	return guardedUpdate(ctx, tx, productID, "reserved_quantity", quantity,
		map[string]interface{}{
			"available_quantity": gorm.Expr("available_quantity + ?", quantity),
			"reserved_quantity":  gorm.Expr("reserved_quantity - ?", quantity),
		},
	)
}

// ConfirmReservedStock deducts reserved stock permanently.
func (r *InventoryRepository) ConfirmReservedStock(
	ctx context.Context,
	productID uuid.UUID,
	quantity int,
) error {

	return r.ConfirmReservedStockTx(ctx, r.db, productID, quantity)
}

// ConfirmReservedStockTx deducts reserved stock permanently inside a transaction.
func (r *InventoryRepository) ConfirmReservedStockTx(
	ctx context.Context,
	tx *gorm.DB,
	productID uuid.UUID,
	quantity int,
) error {

	return guardedUpdate(ctx, tx, productID, "reserved_quantity", quantity,
		map[string]interface{}{
			"reserved_quantity": gorm.Expr("reserved_quantity - ?", quantity),
		},
	)
}

// guardedUpdate applies a stock decrement only while guardColumn still
// holds at least quantity. The check and the write are one statement, so
// concurrent callers cannot drive a quantity negative even without a
// prior row lock.
func guardedUpdate(
	ctx context.Context,
	db *gorm.DB,
	productID uuid.UUID,
	guardColumn string,
	quantity int,
	updates map[string]interface{},
) error {

	result := db.WithContext(ctx).
		Model(&models.Inventory{}).
		Where("product_id = ? AND "+guardColumn+" >= ?", productID, quantity).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrInsufficientQuantity
	}

	return nil
}
