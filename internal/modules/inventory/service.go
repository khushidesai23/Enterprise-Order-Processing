package inventory

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/repository"
)

type Service struct {
	inventoryRepository *repository.InventoryRepository
	productRepository   *repository.ProductRepository
}

func NewService(
	inventoryRepository *repository.InventoryRepository,
	productRepository *repository.ProductRepository,
) *Service {
	return &Service{
		inventoryRepository: inventoryRepository,
		productRepository:   productRepository,
	}
}

// CreateInventory creates inventory for a product.
func (s *Service) CreateInventory(req CreateInventoryRequest) (*InventoryResponse, error) {

	exists, err := s.productRepository.GetByID(req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	if exists == nil {
		return nil, ErrProductNotFound
	}

	inventoryExists, err := s.inventoryRepository.ExistsByProductID(req.ProductID)
	if err != nil {
		return nil, err
	}

	if inventoryExists {
		return nil, ErrInventoryAlreadyExists
	}

	inventory := ToInventoryModel(req)

	if err := s.inventoryRepository.Create(inventory); err != nil {
		return nil, err
	}

	inventory, err = s.inventoryRepository.GetByProductID(req.ProductID)
	if err != nil {
		return nil, err
	}

	response := ToInventoryResponse(inventory)

	return &response, nil
}

// GetInventory returns inventory for a product.
func (s *Service) GetInventory(productID uuid.UUID) (*InventoryResponse, error) {

	inventory, err := s.inventoryRepository.GetByProductID(productID)
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
func (s *Service) GetInventories() ([]InventoryListResponse, error) {

	inventories, err := s.inventoryRepository.GetAll()
	if err != nil {
		return nil, err
	}

	return ToInventoryList(inventories), nil
}

// UpdateInventory updates inventory quantities manually.
func (s *Service) UpdateInventory(
	productID uuid.UUID,
	req UpdateInventoryRequest,
) (*InventoryResponse, error) {

	inventory, err := s.inventoryRepository.GetByProductID(productID)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	if req.AvailableQuantity < 0 || req.ReservedQuantity < 0 {
		return nil, ErrNegativeStock
	}

	UpdateInventoryModel(inventory, req)

	if err := s.inventoryRepository.Update(inventory); err != nil {
		return nil, err
	}

	inventory, err = s.inventoryRepository.GetByProductID(productID)
	if err != nil {
		return nil, err
	}

	response := ToInventoryResponse(inventory)

	return &response, nil
}

// AddStock increases available stock.
func (s *Service) AddStock(
	productID uuid.UUID,
	quantity int,
) (*InventoryResponse, error) {

	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	inventory, err := s.inventoryRepository.GetByProductID(productID)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	if err := s.inventoryRepository.AddStock(productID, quantity); err != nil {
		return nil, err
	}

	inventory, err = s.inventoryRepository.GetByProductID(productID)
	if err != nil {
		return nil, err
	}

	response := ToInventoryResponse(inventory)

	return &response, nil
}

// RemoveStock removes available stock.
func (s *Service) RemoveStock(
	productID uuid.UUID,
	quantity int,
) (*InventoryResponse, error) {

	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	inventory, err := s.inventoryRepository.GetByProductID(productID)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	if inventory.AvailableQuantity < quantity {
		return nil, ErrInsufficientStock
	}

	if err := s.inventoryRepository.RemoveStock(productID, quantity); err != nil {
		return nil, err
	}

	inventory, err = s.inventoryRepository.GetByProductID(productID)
	if err != nil {
		return nil, err
	}

	response := ToInventoryResponse(inventory)

	return &response, nil
}

// ReserveStock reserves stock for an order.
func (s *Service) ReserveStock(
	productID uuid.UUID,
	quantity int,
) (*InventoryResponse, error) {

	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	inventory, err := s.inventoryRepository.GetByProductID(productID)
	if err !=nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	if inventory.AvailableQuantity < quantity {
		return nil, ErrInsufficientStock
	}

	if err := s.inventoryRepository.ReserveStock(productID, quantity); err != nil {
		return nil, err
	}

	inventory, err = s.inventoryRepository.GetByProductID(productID)
	if err != nil {
		return nil, err
	}

	response := ToInventoryResponse(inventory)

	return &response, nil
}

// ReleaseReservedStock returns reserved stock back to available stock.
func (s *Service) ReleaseReservedStock(
	productID uuid.UUID,
	quantity int,
) (*InventoryResponse, error) {

	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	inventory, err := s.inventoryRepository.GetByProductID(productID)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	if inventory.ReservedQuantity < quantity {
		return nil, ErrInsufficientReserved
	}

	if err := s.inventoryRepository.ReleaseReservedStock(productID, quantity); err != nil {
		return nil, err
	}

	inventory, err = s.inventoryRepository.GetByProductID(productID)
	if err != nil {
		return nil, err
	}

	response := ToInventoryResponse(inventory)

	return &response, nil
}

// ConfirmReservedStock deducts reserved stock permanently.
// Called after successful payment.
func (s *Service) ConfirmReservedStock(
	productID uuid.UUID,
	quantity int,
) (*InventoryResponse, error) {

	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	inventory, err := s.inventoryRepository.GetByProductID(productID)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInventoryNotFound
		}

		return nil, err
	}

	if inventory.ReservedQuantity < quantity {
		return nil, ErrInsufficientReserved
	}

	if err := s.inventoryRepository.ConfirmReservedStock(productID, quantity); err != nil {
		return nil, err
	}

	inventory, err = s.inventoryRepository.GetByProductID(productID)
	if err != nil {
		return nil, err
	}

	response := ToInventoryResponse(inventory)

	return &response, nil
}