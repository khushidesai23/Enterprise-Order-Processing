package review

import (
	"github.com/google/uuid"
)

const (
	DefaultPageSize = 10
	MaxPageSize     = 100
)

// Sort orders for a product's reviews.
const (
	SortRecent  = "recent"
	SortHighest = "highest"
	SortLowest  = "lowest"
)

var sortOrders = map[string]string{
	SortRecent:  "created_at DESC",
	SortHighest: "rating DESC, created_at DESC",
	SortLowest:  "rating ASC, created_at DESC",
}

type UpsertReviewRequest struct {
	Rating int    `json:"rating" binding:"required,min=1,max=5"`
	Title  string `json:"title" binding:"max=150"`
	Body   string `json:"body" binding:"max=4000"`
}

type ReviewResponse struct {
	ID               uuid.UUID `json:"id"`
	ProductID        uuid.UUID `json:"product_id"`
	UserID           uuid.UUID `json:"user_id"`
	ReviewerName     string    `json:"reviewer_name"`
	Rating           int       `json:"rating"`
	Title            string    `json:"title"`
	Body             string    `json:"body"`
	VerifiedPurchase bool      `json:"verified_purchase"`
	CreatedAt        string    `json:"created_at"`
	UpdatedAt        string    `json:"updated_at"`
}

// RatingSummary describes all of a product's reviews; Histogram is keyed
// by star rating "1" to "5".
type RatingSummary struct {
	Average   float64          `json:"average"`
	Count     int64            `json:"count"`
	Histogram map[string]int64 `json:"histogram"`
}

type ProductReviewsResponse struct {
	Items   []ReviewResponse `json:"items"`
	Total   int64            `json:"total"`
	Page    int              `json:"page"`
	Limit   int              `json:"limit"`
	Summary RatingSummary    `json:"summary"`
}

// AdminReviewResponse adds the product and reviewer email for moderation.
type AdminReviewResponse struct {
	ReviewResponse
	ProductName   string `json:"product_name"`
	ReviewerEmail string `json:"reviewer_email"`
}

type AdminReviewListResponse struct {
	Items []AdminReviewResponse `json:"items"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}
