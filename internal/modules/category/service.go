package category

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/repository"
)

type Service struct {
	categoryRepository *repository.CategoryRepository
}

func NewService(
	categoryRepository *repository.CategoryRepository,
) *Service {
	return &Service{
		categoryRepository: categoryRepository,
	}
}

// CreateCategory creates a new category.
func (s *Service) CreateCategory(
	ctx context.Context,
	req CreateCategoryRequest,
) (*CategoryResponse, error) {

	exists, err := s.categoryRepository.ExistsByName(
		ctx,
		req.Name,
	)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrCategoryNameExists
	}

	category := ToCategoryModel(req)

	if err := s.categoryRepository.Create(
		ctx,
		category,
	); err != nil {
		return nil, err
	}

	category, err = s.categoryRepository.GetByID(
		ctx,
		category.ID,
	)
	if err != nil {
		return nil, err
	}

	response := ToCategoryResponse(category)

	return &response, nil
}

// GetCategory returns a category by ID.
func (s *Service) GetCategory(
	ctx context.Context,
	id uuid.UUID,
) (*CategoryResponse, error) {

	category, err := s.categoryRepository.GetByID(
		ctx,
		id,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}

		return nil, err
	}

	response := ToCategoryResponse(category)

	return &response, nil
}

// GetCategories returns all categories.
func (s *Service) GetCategories(
	ctx context.Context,
) ([]CategoryListResponse, error) {

	categories, err := s.categoryRepository.GetAll(
		ctx,
	)
	if err != nil {
		return nil, err
	}

	return ToCategoryList(categories), nil
}

// UpdateCategory updates an existing category.
func (s *Service) UpdateCategory(
	ctx context.Context,
	id uuid.UUID,
	req UpdateCategoryRequest,
) (*CategoryResponse, error) {

	category, err := s.categoryRepository.GetByID(
		ctx,
		id,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}

		return nil, err
	}

	exists, err := s.categoryRepository.ExistsByNameExceptID(
		ctx,
		id,
		req.Name,
	)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrCategoryNameExists
	}

	UpdateCategoryModel(category, req)

	if err := s.categoryRepository.Update(
		ctx,
		category,
	); err != nil {
		return nil, err
	}

	category, err = s.categoryRepository.GetByID(
		ctx,
		category.ID,
	)
	if err != nil {
		return nil, err
	}

	response := ToCategoryResponse(category)

	return &response, nil
}

// DeleteCategory deletes a category.
//
// A category cannot be deleted if it is associated with one or more products.
func (s *Service) DeleteCategory(
	ctx context.Context,
	id uuid.UUID,
) error {

	category, err := s.categoryRepository.GetByID(
		ctx,
		id,
	)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}

		return err
	}

	hasProducts, err := s.categoryRepository.HasProducts(
		ctx,
		id,
	)
	if err != nil {
		return err
	}

	if hasProducts {
		return ErrCategoryInUse
	}

	return s.categoryRepository.Delete(
		ctx,
		category,
	)
}