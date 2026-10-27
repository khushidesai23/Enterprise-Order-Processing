package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) *ProductRepository {
	return &ProductRepository{
		db: db,
	}
}

// Create creates a new product.
func (r *ProductRepository) Create(
	ctx context.Context,
	product *models.Product,
) error {
	return r.db.WithContext(ctx).
		Create(product).
		Error
}

// GetByID returns a product by its ID.
// GetByIDs returns the products with the given IDs without associations.
// Missing IDs are simply absent from the result.
func (r *ProductRepository) GetByIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]models.Product, error) {

	var products []models.Product

	err := r.db.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&products).
		Error

	if err != nil {
		return nil, err
	}

	return products, nil
}

func (r *ProductRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*models.Product, error) {

	var product models.Product

	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Inventory").
		First(&product, "id = ?", id).
		Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}

		return nil, err
	}

	return &product, nil
}

// GetAll returns all products.
func (r *ProductRepository) GetAll(
	ctx context.Context,
) ([]models.Product, error) {

	var products []models.Product

	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Inventory").
		Order("created_at DESC").
		Find(&products).
		Error

	if err != nil {
		return nil, err
	}

	return products, nil
}

// GetByCategory returns all products belonging to a category.
func (r *ProductRepository) GetByCategory(
	ctx context.Context,
	categoryID uuid.UUID,
) ([]models.Product, error) {

	var products []models.Product

	err := r.db.WithContext(ctx).
		Preload("Category").
		Preload("Inventory").
		Where("category_id = ?", categoryID).
		Order("created_at DESC").
		Find(&products).
		Error

	if err != nil {
		return nil, err
	}

	return products, nil
}

// ExistsBySKU checks whether a SKU already exists.
func (r *ProductRepository) ExistsBySKU(
	ctx context.Context,
	sku string,
) (bool, error) {

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where("sku = ?", sku).
		Count(&count).
		Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsBySKUExceptID checks duplicate SKU excluding current product.
func (r *ProductRepository) ExistsBySKUExceptID(
	ctx context.Context,
	id uuid.UUID,
	sku string,
) (bool, error) {

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where("sku = ? AND id <> ?", sku, id).
		Count(&count).
		Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CategoryExists checks whether a category exists.
func (r *ProductRepository) CategoryExists(
	ctx context.Context,
	categoryID uuid.UUID,
) (bool, error) {

	var count int64

	err := r.db.WithContext(ctx).
		Model(&models.Category{}).
		Where("id = ?", categoryID).
		Count(&count).
		Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Update updates an existing product.
func (r *ProductRepository) Update(
	ctx context.Context,
	product *models.Product,
) error {

	return r.db.WithContext(ctx).
		Model(&models.Product{}).
		Where("id = ?", product.ID).
		Updates(map[string]interface{}{
			"name":        product.Name,
			"description": product.Description,
			"sku":         product.SKU,
			"price":       product.Price,
			"category_id": product.CategoryID,
		}).
		Error
}

// Delete soft deletes a product.
func (r *ProductRepository) Delete(
	ctx context.Context,
	product *models.Product,
) error {

	return r.db.WithContext(ctx).
		Delete(product).
		Error
}

// ProductSearchFilter narrows GET /products/search. Zero values mean "no
// filter".
type ProductSearchFilter struct {
	Words      []string
	CategoryID *uuid.UUID
	Brands     []string
	MinPrice   *float64
	MaxPrice   *float64
	MinRating  *float64
	InStock    bool
	DealsOnly  bool
}

// facet names a filter that a facet query leaves out.
type facet int

const (
	facetNone facet = iota
	facetCategory
	facetBrand
	facetPrice
)

// searchQuery joins the live category and inventory of each live product
// and applies every filter except the one belonging to skip.
func (r *ProductRepository) searchQuery(ctx context.Context, filter ProductSearchFilter, skip facet) *gorm.DB {
	query := r.db.WithContext(ctx).
		Model(&models.Product{}).
		Joins("JOIN categories c ON c.id = products.category_id AND c.deleted_at IS NULL").
		Joins("LEFT JOIN inventories inv ON inv.product_id = products.id AND inv.deleted_at IS NULL")

	for _, word := range filter.Words {
		pattern := "%" + EscapeLike(word) + "%"
		query = query.Where(
			"(products.name ILIKE ? OR products.brand ILIKE ? OR c.name ILIKE ? OR products.sku ILIKE ?)",
			pattern, pattern, pattern, pattern,
		)
	}

	if filter.CategoryID != nil && skip != facetCategory {
		query = query.Where("products.category_id = ?", *filter.CategoryID)
	}

	if len(filter.Brands) > 0 && skip != facetBrand {
		query = query.Where("products.brand IN ?", filter.Brands)
	}

	if skip != facetPrice {
		if filter.MinPrice != nil {
			query = query.Where("products.price >= ?", *filter.MinPrice)
		}
		if filter.MaxPrice != nil {
			query = query.Where("products.price <= ?", *filter.MaxPrice)
		}
	}

	if filter.MinRating != nil {
		query = query.Where("products.rating_average >= ?", *filter.MinRating)
	}

	if filter.InStock {
		query = query.Where("COALESCE(inv.available_quantity, 0) > 0")
	}

	if filter.DealsOnly {
		query = query.Where("products.mrp > products.price")
	}

	return query
}

// Search returns one page of matching products, ordered by orderBy, and
// the total number of matches.
func (r *ProductRepository) Search(
	ctx context.Context,
	filter ProductSearchFilter,
	orderBy string,
	orderVars []any,
	joinSales bool,
	limit int,
	offset int,
) ([]models.Product, int64, error) {

	var total int64
	if err := r.searchQuery(ctx, filter, facetNone).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := r.searchQuery(ctx, filter, facetNone)
	if joinSales {
		query = query.Joins(`LEFT JOIN (
			SELECT oi.product_id, SUM(oi.quantity) AS sold
			FROM order_items oi
			JOIN orders o ON o.id = oi.order_id AND o.deleted_at IS NULL
			WHERE oi.deleted_at IS NULL AND o.status IN ('PAID', 'PACKED', 'SHIPPED', 'DELIVERED')
			GROUP BY oi.product_id
		) sales ON sales.product_id = products.id`)
	}

	var products []models.Product
	err := query.
		Select("products.*").
		Preload("Category").
		Preload("Inventory").
		// One ORDER BY expression: GORM drops an expression ORDER BY when
		// another Order() is merged into it. products.id keeps pages stable.
		Order(clause.OrderBy{Expression: clause.Expr{SQL: orderBy + ", products.id", Vars: orderVars, WithoutParentheses: true}}).
		Limit(limit).
		Offset(offset).
		Find(&products).
		Error
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

type CategoryCount struct {
	ID    uuid.UUID
	Name  string
	Count int64
}

type BrandCount struct {
	Name  string
	Count int64
}

// SearchFacets counts matches per category and per brand, and the price
// range, each ignoring its own filter.
func (r *ProductRepository) SearchFacets(
	ctx context.Context,
	filter ProductSearchFilter,
	maxBrands int,
) ([]CategoryCount, []BrandCount, float64, float64, error) {

	var categories []CategoryCount
	err := r.searchQuery(ctx, filter, facetCategory).
		Select("c.id AS id, c.name AS name, COUNT(*) AS count").
		Group("c.id, c.name").
		Order("count DESC, c.name").
		Scan(&categories).
		Error
	if err != nil {
		return nil, nil, 0, 0, err
	}

	var brands []BrandCount
	err = r.searchQuery(ctx, filter, facetBrand).
		Where("products.brand <> ''").
		Select("products.brand AS name, COUNT(*) AS count").
		Group("products.brand").
		Order("count DESC, products.brand").
		Limit(maxBrands).
		Scan(&brands).
		Error
	if err != nil {
		return nil, nil, 0, 0, err
	}

	var prices struct {
		Min float64
		Max float64
	}
	err = r.searchQuery(ctx, filter, facetPrice).
		Select("COALESCE(MIN(products.price), 0) AS min, COALESCE(MAX(products.price), 0) AS max").
		Scan(&prices).
		Error
	if err != nil {
		return nil, nil, 0, 0, err
	}

	return categories, brands, prices.Min, prices.Max, nil
}

// EscapeLike makes user text match literally inside LIKE/ILIKE.
func EscapeLike(value string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(value)
}
