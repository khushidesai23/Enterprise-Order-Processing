package product

import (
	"github.com/google/uuid"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

type CreateProductRequest struct {
	Name        string    `json:"name" binding:"required,min=2,max=255"`
	Description string    `json:"description" binding:"max=1000"`
	SKU         string    `json:"sku" binding:"required,min=2,max=100"`
	Price       float64   `json:"price" binding:"required,gt=0"`
	CategoryID  uuid.UUID `json:"category_id" binding:"required"`

	// Optional catalogue details.
	Brand    string               `json:"brand" binding:"max=100"`
	MRP      float64              `json:"mrp" binding:"gte=0"`
	ImageURL string               `json:"image_url" binding:"max=500"`
	Images   []string             `json:"images" binding:"max=10,dive,max=500"`
	Specs    []models.ProductSpec `json:"specs" binding:"max=30"`
}

// UpdateProductRequest replaces the core fields. The optional catalogue
// details are left unchanged when omitted (nil), so older clients that do
// not know them do not erase them.
type UpdateProductRequest struct {
	Name        string    `json:"name" binding:"required,min=2,max=255"`
	Description string    `json:"description" binding:"max=1000"`
	SKU         string    `json:"sku" binding:"required,min=2,max=100"`
	Price       float64   `json:"price" binding:"required,gt=0"`
	CategoryID  uuid.UUID `json:"category_id" binding:"required"`

	Brand    *string               `json:"brand" binding:"omitempty,max=100"`
	MRP      *float64              `json:"mrp" binding:"omitempty,gte=0"`
	ImageURL *string               `json:"image_url" binding:"omitempty,max=500"`
	Images   *[]string             `json:"images" binding:"omitempty,max=10,dive,max=500"`
	Specs    *[]models.ProductSpec `json:"specs" binding:"omitempty,max=30"`
}

type ProductResponse struct {
	ID            uuid.UUID            `json:"id"`
	Name          string               `json:"name"`
	Description   string               `json:"description"`
	SKU           string               `json:"sku"`
	Price         float64              `json:"price"`
	Brand         string               `json:"brand"`
	MRP           float64              `json:"mrp"`
	ImageURL      string               `json:"image_url"`
	Images        []string             `json:"images"`
	Specs         []models.ProductSpec `json:"specs"`
	RatingAverage float64              `json:"rating_average"`
	RatingCount   int                  `json:"rating_count"`
	Category      CategoryResponse     `json:"category"`
	Inventory     InventorySummary     `json:"inventory,omitempty"`
}

type ProductListResponse struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	SKU           string    `json:"sku"`
	Price         float64   `json:"price"`
	Brand         string    `json:"brand"`
	MRP           float64   `json:"mrp"`
	ImageURL      string    `json:"image_url"`
	RatingAverage float64   `json:"rating_average"`
	RatingCount   int       `json:"rating_count"`
	CategoryID    uuid.UUID `json:"category_id"`
	Category      string    `json:"category"`
	Available     int       `json:"available_quantity"`
	Reserved      int       `json:"reserved_quantity"`
}

type CategoryResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type InventorySummary struct {
	AvailableQuantity int `json:"available_quantity"`
	ReservedQuantity  int `json:"reserved_quantity"`
}

type ProductStatistics struct {
	TotalProducts int64 `json:"total_products"`
	ActiveSKU     int64 `json:"active_sku"`
}

type ProductWithRelations struct {
	Product models.Product
}

// Search

const (
	DefaultSearchPageSize = 24
	MaxSearchPageSize     = 100
)

// Sort orders accepted by GET /products/search.
const (
	SortRelevance   = "relevance"
	SortFeatured    = "featured"
	SortPriceAsc    = "price_asc"
	SortPriceDesc   = "price_desc"
	SortRating      = "rating"
	SortReviews     = "reviews"
	SortNewest      = "newest"
	SortDiscount    = "discount"
	SortBestselling = "bestselling"
)

var searchSorts = map[string]bool{
	SortRelevance: true, SortFeatured: true, SortPriceAsc: true, SortPriceDesc: true,
	SortRating: true, SortReviews: true, SortNewest: true, SortDiscount: true, SortBestselling: true,
}

type SearchProductsRequest struct {
	Query      string
	CategoryID *uuid.UUID
	Brands     []string
	MinPrice   *float64
	MaxPrice   *float64
	MinRating  *float64
	InStock    bool
	DealsOnly  bool
	Sort       string
	Page       int
	Limit      int
}

type ProductSearchResponse struct {
	Items  []ProductListResponse `json:"items"`
	Total  int64                 `json:"total"`
	Page   int                   `json:"page"`
	Limit  int                   `json:"limit"`
	Facets SearchFacets          `json:"facets"`
}

// SearchFacets describe the result set for filter UIs. Each facet ignores
// its own filter, so choosing one brand still lists the other brands.
type SearchFacets struct {
	Categories []CategoryFacet `json:"categories"`
	Brands     []BrandFacet    `json:"brands"`
	PriceMin   float64         `json:"price_min"`
	PriceMax   float64         `json:"price_max"`
}

type CategoryFacet struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Count int64     `json:"count"`
}

type BrandFacet struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}
