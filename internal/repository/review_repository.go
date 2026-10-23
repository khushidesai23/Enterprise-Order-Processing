package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

type ReviewRepository struct {
	db *gorm.DB
}

func NewReviewRepository(db *gorm.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

func (r *ReviewRepository) Begin(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Begin()
}

func (r *ReviewRepository) ProductExists(ctx context.Context, productID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Product{}).Where("id = ?", productID).Count(&count).Error
	return count > 0, err
}

// LockProductTx locks the product row so concurrent review changes
// recompute its rating one at a time.
func (r *ReviewRepository) LockProductTx(tx *gorm.DB, productID uuid.UUID) error {
	var product models.Product
	return tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Select("id").
		First(&product, "id = ?", productID).
		Error
}

// GetByProductAndUserTx returns the user's live review of a product, or
// nil when there is none.
func (r *ReviewRepository) GetByProductAndUserTx(tx *gorm.DB, productID, userID uuid.UUID) (*models.Review, error) {
	var review models.Review
	err := tx.Where("product_id = ? AND user_id = ?", productID, userID).First(&review).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *ReviewRepository) GetByProductAndUser(ctx context.Context, productID, userID uuid.UUID) (*models.Review, error) {
	return r.GetByProductAndUserTx(r.db.WithContext(ctx).Preload("User"), productID, userID)
}

func (r *ReviewRepository) SaveTx(tx *gorm.DB, review *models.Review) error {
	return tx.Omit(clause.Associations).Save(review).Error
}

func (r *ReviewRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Review, error) {
	var review models.Review
	err := r.db.WithContext(ctx).Preload("User").First(&review, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &review, nil
}

func (r *ReviewRepository) DeleteTx(tx *gorm.DB, review *models.Review) error {
	return tx.Delete(review).Error
}

// RefreshProductRatingTx recomputes the product's average rating and
// review count from its live reviews.
func (r *ReviewRepository) RefreshProductRatingTx(tx *gorm.DB, productID uuid.UUID) error {
	return tx.Exec(`
		UPDATE products SET
			rating_average = COALESCE((SELECT ROUND(AVG(rating)::numeric, 2) FROM reviews WHERE product_id = @id AND deleted_at IS NULL), 0),
			rating_count = (SELECT COUNT(*) FROM reviews WHERE product_id = @id AND deleted_at IS NULL)
		WHERE id = @id`,
		map[string]any{"id": productID},
	).Error
}

// HasPurchased reports whether the user has a paid order containing the
// product.
func (r *ReviewRepository) HasPurchased(ctx context.Context, userID, productID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("order_items oi").
		Joins("JOIN orders o ON o.id = oi.order_id AND o.deleted_at IS NULL").
		Where("oi.deleted_at IS NULL AND oi.product_id = ? AND o.user_id = ?", productID, userID).
		Where("o.status IN ?", []models.OrderStatus{models.OrderPaid, models.OrderPacked, models.OrderShipped, models.OrderDelivered}).
		Count(&count).
		Error
	return count > 0, err
}

// ListByProduct returns one page of a product's reviews and their total.
func (r *ReviewRepository) ListByProduct(
	ctx context.Context,
	productID uuid.UUID,
	rating int,
	orderBy string,
	limit int,
	offset int,
) ([]models.Review, int64, error) {

	query := r.db.WithContext(ctx).Model(&models.Review{}).Where("product_id = ?", productID)
	if rating > 0 {
		query = query.Where("rating = ?", rating)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var reviews []models.Review
	err := query.Preload("User").Order(orderBy).Order("id").Limit(limit).Offset(offset).Find(&reviews).Error
	if err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

// RatingHistogram counts a product's reviews per star rating.
func (r *ReviewRepository) RatingHistogram(ctx context.Context, productID uuid.UUID) (map[int]int64, error) {
	var rows []struct {
		Rating int
		Count  int64
	}
	err := r.db.WithContext(ctx).
		Model(&models.Review{}).
		Select("rating, COUNT(*) AS count").
		Where("product_id = ?", productID).
		Group("rating").
		Scan(&rows).
		Error
	if err != nil {
		return nil, err
	}

	histogram := make(map[int]int64, len(rows))
	for _, row := range rows {
		histogram[row.Rating] = row.Count
	}
	return histogram, nil
}

// ListAll returns one page of all reviews, newest first, for moderation.
func (r *ReviewRepository) ListAll(ctx context.Context, rating int, limit, offset int) ([]models.Review, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Review{})
	if rating > 0 {
		query = query.Where("rating = ?", rating)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var reviews []models.Review
	err := query.Preload("User").Preload("Product").
		Order("created_at DESC").Order("id").
		Limit(limit).Offset(offset).
		Find(&reviews).Error
	if err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}
