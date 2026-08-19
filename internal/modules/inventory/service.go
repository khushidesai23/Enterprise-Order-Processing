package inventory

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/logger"
)

type Service struct {
	inventoryRepository inventoryRepository
	productRepository   inventoryProductRepository
	log                 *zap.Logger
}

type inventoryRepository interface {
	Create(ctx context.Context, inventory *models.Inventory) error
	GetByProductID(ctx context.Context, productID uuid.UUID) (*models.Inventory, error)
	GetAll(ctx context.Context) ([]models.Inventory, error)
	ExistsByProductID(ctx context.Context, productID uuid.UUID) (bool, error)
	Update(ctx context.Context, inv *models.Inventory) error
	AddStock(ctx context.Context, productID uuid.UUID, quantity int) error
	RemoveStock(ctx context.Context, productID uuid.UUID, quantity int) error
	ReserveStock(ctx context.Context, productID uuid.UUID, quantity int) error
	ReleaseReservedStock(ctx context.Context, productID uuid.UUID, quantity int) error
	ConfirmReservedStock(ctx context.Context, productID uuid.UUID, quantity int) error
}

type inventoryProductRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
}

func NewService(
	inventoryRepository inventoryRepository,
	productRepository inventoryProductRepository,
	log *zap.Logger,
) *Service {
	return &Service{
		inventoryRepository: inventoryRepository,
		productRepository:   productRepository,
		log:                 log,
	}
}

// CreateInventory creates inventory for a product.
func (s *Service) CreateInventory(
	ctx context.Context,
	req CreateInventoryRequest,
) (*InventoryResponse, error) {

	product, err := s.productRepository.GetByID(
		ctx,
		req.ProductID,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}

		return nil, err
	}

	if product == nil {
		return nil, ErrProductNotFound
	}

	inventoryExists, err := s.inventoryRepository.ExistsByProductID(
		ctx,
		req.ProductID,
	)
	if err != nil {
		return nil, err
	}

	if inventoryExists {
		return nil, ErrInventoryAlreadyExists
	}

	inventory := ToInventoryModel(req)

	if err := s.inventoryRepository.Create(
		ctx,
		inventory,
	); err != nil {
		return nil, err
	}

	inventory, err = s.inventoryRepository.GetByProductID(
		ctx,
		req.ProductID,
	)
	if err != nil {
		return nil, err
	}

	logger.InfoContext(
		ctx,
		s.log,
		"inventory created",
		zap.String("product_id", req.ProductID.String()),
	)

	response := ToInventoryResponse(inventory)

	return &response, nil
}

// GetInventory returns inventory for a product.
func (s *Service) GetInventory(
	ctx context.Context,
	productID uuid.UUID,
) (*InventoryResponse, error) {

	inventory, err := s.inventoryRepository.GetByProductID(
		ctx,
		productID,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	response := ToInventoryResponse(inventory)

	return &response, nil
}

// GetInventories returns all inventory records.
func (s *Service) GetInventories(
	ctx context.Context,
) ([]InventoryListResponse, error) {

	inventories, err := s.inventoryRepository.GetAll(
		ctx,
	)
	if err != nil {
		return nil, err
	}

	return ToInventoryList(inventories), nil
}

// UpdateInventory updates inventory quantities manually.
func (s *Service) UpdateInventory(
	ctx context.Context,
	productID uuid.UUID,
	req UpdateInventoryRequest,
) (*InventoryResponse, error) {

	inventory, err := s.inventoryRepository.GetByProductID(
		ctx,
		productID,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	if req.AvailableQuantity < 0 ||
		req.ReservedQuantity < 0 {
		return nil, ErrNegativeStock
	}

	UpdateInventoryModel(
		inventory,
		req,
	)

	if err := s.inventoryRepository.Update(
		ctx,
		inventory,
	); err != nil {
		return nil, err
	}

	inventory, err = s.inventoryRepository.GetByProductID(
		ctx,
		productID,
	)
	if err != nil {
		return nil, err
	}

	logger.InfoContext(
		ctx,
		s.log,
		"inventory updated",
		zap.String("product_id", productID.String()),
	)

	response := ToInventoryResponse(inventory)

	return &response, nil
}

// AddStock increases available stock.
func (s *Service) AddStock(
	ctx context.Context,
	productID uuid.UUID,
	quantity int,
) (*InventoryResponse, error) {

	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	inventory, err := s.inventoryRepository.GetByProductID(
		ctx,
		productID,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	if err := s.inventoryRepository.AddStock(
		ctx,
		productID,
		quantity,
	); err != nil {
		return nil, err
	}

	inventory, err = s.inventoryRepository.GetByProductID(
		ctx,
		productID,
	)
	if err != nil {
		return nil, err
	}

	response := ToInventoryResponse(inventory)

	return &response, nil
}

// RemoveStock removes available stock.
func (s *Service) RemoveStock(
	ctx context.Context,
	productID uuid.UUID,
	quantity int,
) (*InventoryResponse, error) {

	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	inventory, err := s.inventoryRepository.GetByProductID(
		ctx,
		productID,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	if inventory.AvailableQuantity < quantity {
		return nil, ErrInsufficientStock
	}

	if err := s.inventoryRepository.RemoveStock(
		ctx,
		productID,
		quantity,
	); err != nil {
		return nil, err
	}

	inventory, err = s.inventoryRepository.GetByProductID(
		ctx,
		productID,
	)
	if err != nil {
		return nil, err
	}

	response := ToInventoryResponse(inventory)

	return &response, nil
}

// ReserveStock reserves stock for an order.
func (s *Service) ReserveStock(
	ctx context.Context,
	productID uuid.UUID,
	quantity int,
) (*InventoryResponse, error) {

	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	inventory, err := s.inventoryRepository.GetByProductID(
		ctx,
		productID,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	if inventory.AvailableQuantity < quantity {
		return nil, ErrInsufficientStock
	}

	if err := s.inventoryRepository.ReserveStock(
		ctx,
		productID,
		quantity,
	); err != nil {
		return nil, err
	}

	inventory, err = s.inventoryRepository.GetByProductID(
		ctx,
		productID,
	)
	if err != nil {
		return nil, err
	}

	logger.InfoContext(
		ctx,
		s.log,
		"inventory stock reserved",
		zap.String("product_id", productID.String()),
		zap.Int("quantity", quantity),
	)

	response := ToInventoryResponse(inventory)

	return &response, nil
}

// ReleaseReservedStock returns reserved stock back to available stock.
func (s *Service) ReleaseReservedStock(
	ctx context.Context,
	productID uuid.UUID,
	quantity int,
) (*InventoryResponse, error) {

	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	inventory, err := s.inventoryRepository.GetByProductID(
		ctx,
		productID,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	if inventory.ReservedQuantity < quantity {
		return nil, ErrInsufficientReserved
	}

	if err := s.inventoryRepository.ReleaseReservedStock(
		ctx,
		productID,
		quantity,
	); err != nil {
		return nil, err
	}

	inventory, err = s.inventoryRepository.GetByProductID(
		ctx,
		productID,
	)
	if err != nil {
		return nil, err
	}

	logger.InfoContext(
		ctx,
		s.log,
		"inventory stock released",
		zap.String("product_id", productID.String()),
		zap.Int("quantity", quantity),
	)

	response := ToInventoryResponse(inventory)

	return &response, nil
}

// ConfirmReservedStock deducts reserved stock permanently.
// Called after successful payment.
func (s *Service) ConfirmReservedStock(
	ctx context.Context,
	productID uuid.UUID,
	quantity int,
) (*InventoryResponse, error) {

	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	inventory, err := s.inventoryRepository.GetByProductID(
		ctx,
		productID,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	if inventory.ReservedQuantity < quantity {
		return nil, ErrInsufficientReserved
	}

	if err := s.inventoryRepository.ConfirmReservedStock(
		ctx,
		productID,
		quantity,
	); err != nil {
		return nil, err
	}

	inventory, err = s.inventoryRepository.GetByProductID(
		ctx,
		productID,
	)
	if err != nil {
		return nil, err
	}

	logger.InfoContext(
		ctx,
		s.log,
		"inventory reserved stock confirmed",
		zap.String("product_id", productID.String()),
		zap.Int("quantity", quantity),
	)

	response := ToInventoryResponse(inventory)

	return &response, nil
}
