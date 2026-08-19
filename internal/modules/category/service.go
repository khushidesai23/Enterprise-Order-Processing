package category

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/logger"
)

type CategoryRepository interface {
	Create(ctx context.Context, category *models.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Category, error)
	GetAll(ctx context.Context) ([]models.Category, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
	ExistsByNameExceptID(ctx context.Context, id uuid.UUID, name string) (bool, error)
	HasProducts(ctx context.Context, id uuid.UUID) (bool, error)
	Update(ctx context.Context, category *models.Category) error
	Delete(ctx context.Context, category *models.Category) error
}

type Service struct {
	categoryRepository CategoryRepository
	log                *zap.Logger
}

func NewService(
	categoryRepository CategoryRepository,
	log *zap.Logger,
) *Service {
	return &Service{
		categoryRepository: categoryRepository,
		log:                log,
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

	logger.InfoContext(
		ctx,
		s.log,
		"category created",
		zap.String("category_id", category.ID.String()),
		zap.String("name", category.Name),
	)

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

	logger.InfoContext(
		ctx,
		s.log,
		"category updated",
		zap.String("category_id", category.ID.String()),
		zap.String("name", category.Name),
	)

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

	if err := s.categoryRepository.Delete(ctx, category); err != nil {
		return err
	}

	logger.InfoContext(
		ctx,
		s.log,
		"category deleted",
		zap.String("category_id", category.ID.String()),
		zap.String("name", category.Name),
	)

	return nil
}
