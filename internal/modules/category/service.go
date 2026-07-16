package category

import (
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
func (s *Service) CreateCategory(req CreateCategoryRequest) (*CategoryResponse, error) {

	exists, err := s.categoryRepository.ExistsByName(req.Name)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrCategoryNameExists
	}

	category := ToCategoryModel(req)

	if err := s.categoryRepository.Create(category); err != nil {
		return nil, err
	}

	category, err = s.categoryRepository.GetByID(category.ID)
	if err != nil {
		return nil, err
	}

	response := ToCategoryResponse(category)

	return &response, nil
}

// GetCategory returns a category by ID.
func (s *Service) GetCategory(id uuid.UUID) (*CategoryResponse, error) {

	category, err := s.categoryRepository.GetByID(id)
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
func (s *Service) GetCategories() ([]CategoryListResponse, error) {

	categories, err := s.categoryRepository.GetAll()
	if err != nil {
		return nil, err
	}

	return ToCategoryList(categories), nil
}

// UpdateCategory updates an existing category.
func (s *Service) UpdateCategory(
	id uuid.UUID,
	req UpdateCategoryRequest,
) (*CategoryResponse, error) {

	category, err := s.categoryRepository.GetByID(id)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCategoryNotFound
		}

		return nil, err
	}

	exists, err := s.categoryRepository.ExistsByNameExceptID(id, req.Name)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, ErrCategoryNameExists
	}

	UpdateCategoryModel(category, req)

	if err := s.categoryRepository.Update(category); err != nil {
		return nil, err
	}

	category, err = s.categoryRepository.GetByID(category.ID)
	if err != nil {
		return nil, err
	}

	response := ToCategoryResponse(category)

	return &response, nil
}

// DeleteCategory deletes a category.
//
// A category cannot be deleted if it is associated with one or more products.
func (s *Service) DeleteCategory(id uuid.UUID) error {

	category, err := s.categoryRepository.GetByID(id)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}

		return err
	}

	hasProducts, err := s.categoryRepository.HasProducts(id)
	if err != nil {
		return err
	}

	if hasProducts {
		return ErrCategoryInUse
	}

	return s.categoryRepository.Delete(category)
}