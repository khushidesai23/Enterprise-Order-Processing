package models

import "github.com/google/uuid"

type Product struct {
	BaseModel

	Name        string  `gorm:"size:255;not null"`
	Description string  `gorm:"type:text"`
	SKU         string  `gorm:"size:100;uniqueIndex:idx_products_sku,where:deleted_at IS NULL;not null"`
	Price       float64 `gorm:"type:numeric(12,2);not null"`

	// Brand is optional; it drives brand filters in search.
	Brand string `gorm:"size:100;index;not null;default:''"`

	// MRP is the list price shown struck through next to Price; 0 when
	// the product is not discounted.
	MRP float64 `gorm:"type:numeric(12,2);not null;default:0"`

	// ImageURL is the main image; Images holds the full gallery.
	ImageURL string                `gorm:"size:500;not null;default:''"`
	Images   JSONList[string]      `gorm:"not null;default:'[]'"`
	Specs    JSONList[ProductSpec] `gorm:"not null;default:'[]'"`

	// Maintained by the review module whenever a review changes.
	RatingAverage float64 `gorm:"type:numeric(3,2);not null;default:0"`
	RatingCount   int     `gorm:"not null;default:0"`

	CategoryID uuid.UUID
	Category   Category `gorm:"foreignKey:CategoryID"`

	Inventory *Inventory `gorm:"foreignKey:ProductID"`
}

// ProductSpec is one row of a product's specification table.
type ProductSpec struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
