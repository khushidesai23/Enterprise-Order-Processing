package product

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/repository"
)

type Service struct {
	productRepository *repository.ProductRepository
}

func NewService(
	productRepository *repository.ProductRepository,
) *Service {
	return &Service{
		productRepository: productRepository,
	}
}

// CreateProduct creates a new product.
func (s *Service) CreateProduct(
	ctx context.Context,
	req CreateProductRequest,
) (*ProductResponse, error) {

	// Check Category
	exists, err := s.productRepository.CategoryExists(
		ctx,
		req.CategoryID,
	)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrCategoryNotFound
	}

	// Check SKU
	exists, err = s.productRepository.ExistsBySKU(
		ctx,
		req.SKU,
	)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrProductSKUExists
	}

	product := ToProductModel(req)

	if err := s.productRepository.Create(
		ctx,
		product,
	); err != nil {
		return nil, err
	}

	product, err = s.productRepository.GetByID(
		ctx,
		product.ID,
	)
	if err != nil {
		return nil, err
	}

	response := ToProductResponse(product)

	return &response, nil
}

// GetProduct returns a product by ID.
func (s *Service) GetProduct(
	ctx context.Context,
	id uuid.UUID,
) (*ProductResponse, error) {

	product, err := s.productRepository.GetByID(
		ctx,
		id,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}

		return nil, err
	}

	response := ToProductResponse(product)

	return &response, nil
}

// GetProducts returns all products.
func (s *Service) GetProducts(
	ctx context.Context,
) ([]ProductListResponse, error) {

	products, err := s.productRepository.GetAll(
		ctx,
	)
	if err != nil {
		return nil, err
	}

	return ToProductList(products), nil
}

// GetProductsByCategory returns products of a category.
func (s *Service) GetProductsByCategory(
	ctx context.Context,
	categoryID uuid.UUID,
) ([]ProductListResponse, error) {

	exists, err := s.productRepository.CategoryExists(
		ctx,
		categoryID,
	)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrCategoryNotFound
	}

	products, err := s.productRepository.GetByCategory(
		ctx,
		categoryID,
	)
	if err != nil {
		return nil, err
	}

	return ToProductList(products), nil
}

// UpdateProduct updates an existing product.
func (s *Service) UpdateProduct(
	ctx context.Context,
	id uuid.UUID,
	req UpdateProductRequest,
) (*ProductResponse, error) {

	product, err := s.productRepository.GetByID(
		ctx,
		id,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}

		return nil, err
	}

	exists, err := s.productRepository.CategoryExists(
		ctx,
		req.CategoryID,
	)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, ErrCategoryNotFound
	}

	exists, err = s.productRepository.ExistsBySKUExceptID(
		ctx,
		id,
		req.SKU,
	)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrProductSKUExists
	}

	UpdateProductModel(product, req)

	if err := s.productRepository.Update(
		ctx,
		product,
	); err != nil {
		return nil, err
	}

	product, err = s.productRepository.GetByID(
		ctx,
		product.ID,
	)
	if err != nil {
		return nil, err
	}

	response := ToProductResponse(product)

	return &response, nil
}

// DeleteProduct soft deletes a product.
func (s *Service) DeleteProduct(
	ctx context.Context,
	id uuid.UUID,
) error {

	product, err := s.productRepository.GetByID(
		ctx,
		id,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductNotFound
		}

		return err
	}

	return s.productRepository.Delete(
		ctx,
		product,
	)
}