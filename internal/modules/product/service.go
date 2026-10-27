package product

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/repository"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/logger"
)

type Service struct {
	productRepository productRepository
	log               *zap.Logger
}

type productRepository interface {
	Create(ctx context.Context, product *models.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Product, error)
	GetAll(ctx context.Context) ([]models.Product, error)
	GetByCategory(ctx context.Context, categoryID uuid.UUID) ([]models.Product, error)
	ExistsBySKU(ctx context.Context, sku string) (bool, error)
	ExistsBySKUExceptID(ctx context.Context, id uuid.UUID, sku string) (bool, error)
	CategoryExists(ctx context.Context, categoryID uuid.UUID) (bool, error)
	Update(ctx context.Context, product *models.Product) error
	Delete(ctx context.Context, product *models.Product) error
	Search(ctx context.Context, filter repository.ProductSearchFilter, orderBy string, orderVars []any, joinSales bool, limit, offset int) ([]models.Product, int64, error)
	SearchFacets(ctx context.Context, filter repository.ProductSearchFilter, maxBrands int) ([]repository.CategoryCount, []repository.BrandCount, float64, float64, error)
}

func NewService(
	productRepository productRepository,
	log *zap.Logger,
) *Service {
	return &Service{
		productRepository: productRepository,
		log:               log,
	}
}

// CreateProduct creates a new product.
func (s *Service) CreateProduct(
	ctx context.Context,
	req CreateProductRequest,
) (*ProductResponse, error) {

	if req.MRP > 0 && req.MRP < req.Price {
		return nil, ErrInvalidMRP
	}

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

	logger.InfoContext(
		ctx,
		s.log,
		"product created",
		zap.String("product_id", product.ID.String()),
		zap.String("sku", product.SKU),
	)

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

	if product.MRP > 0 && product.MRP < product.Price {
		return nil, ErrInvalidMRP
	}

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

	logger.InfoContext(
		ctx,
		s.log,
		"product updated",
		zap.String("product_id", product.ID.String()),
		zap.String("sku", product.SKU),
	)

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

	if err := s.productRepository.Delete(ctx, product); err != nil {
		return err
	}

	logger.InfoContext(
		ctx,
		s.log,
		"product deleted",
		zap.String("product_id", product.ID.String()),
	)

	return nil
}

// searchOrders maps each sort to its ORDER BY clause.
var searchOrders = map[string]string{
	SortFeatured:    "products.rating_count DESC, products.rating_average DESC",
	SortPriceAsc:    "products.price ASC",
	SortPriceDesc:   "products.price DESC",
	SortRating:      "products.rating_average DESC, products.rating_count DESC",
	SortReviews:     "products.rating_count DESC",
	SortNewest:      "products.created_at DESC",
	SortDiscount:    "CASE WHEN products.mrp > products.price THEN (products.mrp - products.price) / products.mrp ELSE 0 END DESC",
	SortBestselling: "COALESCE(sales.sold, 0) DESC, products.rating_count DESC",
}

const maxBrandFacets = 40

// SearchProducts returns one page of products matching the request, with
// facets describing the whole result set.
func (s *Service) SearchProducts(
	ctx context.Context,
	req SearchProductsRequest,
) (*ProductSearchResponse, error) {

	filter := repository.ProductSearchFilter{
		Words:      strings.Fields(req.Query),
		CategoryID: req.CategoryID,
		Brands:     req.Brands,
		MinPrice:   req.MinPrice,
		MaxPrice:   req.MaxPrice,
		MinRating:  req.MinRating,
		InStock:    req.InStock,
		DealsOnly:  req.DealsOnly,
	}

	sort := req.Sort
	if sort == "" || sort == SortRelevance && len(filter.Words) == 0 {
		sort = SortFeatured
		if len(filter.Words) > 0 {
			sort = SortRelevance
		}
	}

	orderBy, ok := searchOrders[sort]
	var orderVars []any
	if sort == SortRelevance {
		// Names that start with the query, then names that contain it,
		// then matches on brand, category or SKU only.
		phrase := repository.EscapeLike(strings.Join(filter.Words, " "))
		orderBy = "CASE WHEN products.name ILIKE ? THEN 0 WHEN products.name ILIKE ? THEN 1 ELSE 2 END, products.rating_count DESC"
		orderVars = []any{phrase + "%", "%" + phrase + "%"}
		ok = true
	}
	if !ok {
		return nil, ErrInvalidSort
	}

	products, total, err := s.productRepository.Search(
		ctx,
		filter,
		orderBy,
		orderVars,
		sort == SortBestselling,
		req.Limit,
		(req.Page-1)*req.Limit,
	)
	if err != nil {
		return nil, err
	}

	categories, brands, priceMin, priceMax, err := s.productRepository.SearchFacets(ctx, filter, maxBrandFacets)
	if err != nil {
		return nil, err
	}

	response := ProductSearchResponse{
		Items: ToProductList(products),
		Total: total,
		Page:  req.Page,
		Limit: req.Limit,
		Facets: SearchFacets{
			Categories: make([]CategoryFacet, 0, len(categories)),
			Brands:     make([]BrandFacet, 0, len(brands)),
			PriceMin:   priceMin,
			PriceMax:   priceMax,
		},
	}

	for _, category := range categories {
		response.Facets.Categories = append(response.Facets.Categories, CategoryFacet{
			ID: category.ID, Name: category.Name, Count: category.Count,
		})
	}

	for _, brand := range brands {
		response.Facets.Brands = append(response.Facets.Brands, BrandFacet{
			Name: brand.Name, Count: brand.Count,
		})
	}

	return &response, nil
}
