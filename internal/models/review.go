package models

import "github.com/google/uuid"

// Review is a customer's rating of a product. A user has at most one live
// review per product.
type Review struct {
	BaseModel

	ProductID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_reviews_product_user,where:deleted_at IS NULL"`
	Product   Product   `gorm:"foreignKey:ProductID"`

	UserID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_reviews_product_user,where:deleted_at IS NULL"`
	User   User      `gorm:"foreignKey:UserID"`

	Rating int    `gorm:"not null;check:rating BETWEEN 1 AND 5"`
	Title  string `gorm:"size:150;not null;default:''"`
	Body   string `gorm:"type:text;not null;default:''"`

	// VerifiedPurchase is set when the reviewer has a paid order that
	// contains the product.
	VerifiedPurchase bool `gorm:"not null;default:false"`
}
