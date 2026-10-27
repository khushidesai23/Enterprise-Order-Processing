package review

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/authz"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/logger"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrReviewNotFound  = errors.New("review not found")
	ErrInvalidSort     = errors.New("sort must be recent, highest or lowest")
)

type reviewRepository interface {
	Begin(ctx context.Context) *gorm.DB
	ProductExists(ctx context.Context, productID uuid.UUID) (bool, error)
	LockProductTx(tx *gorm.DB, productID uuid.UUID) error
	GetByProductAndUserTx(tx *gorm.DB, productID, userID uuid.UUID) (*models.Review, error)
	GetByProductAndUser(ctx context.Context, productID, userID uuid.UUID) (*models.Review, error)
	SaveTx(tx *gorm.DB, review *models.Review) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Review, error)
	DeleteTx(tx *gorm.DB, review *models.Review) error
	RefreshProductRatingTx(tx *gorm.DB, productID uuid.UUID) error
	HasPurchased(ctx context.Context, userID, productID uuid.UUID) (bool, error)
	ListByProduct(ctx context.Context, productID uuid.UUID, rating int, orderBy string, limit, offset int) ([]models.Review, int64, error)
	RatingHistogram(ctx context.Context, productID uuid.UUID) (map[int]int64, error)
	ListAll(ctx context.Context, rating int, limit, offset int) ([]models.Review, int64, error)
}

type Service struct {
	repository reviewRepository
	log        *zap.Logger
}

func NewService(repository reviewRepository, log *zap.Logger) *Service {
	return &Service{repository: repository, log: log}
}

// ListProductReviews returns one page of a product's reviews (optionally
// only one star rating) with the summary of all its reviews.
func (s *Service) ListProductReviews(
	ctx context.Context,
	productID uuid.UUID,
	rating int,
	sort string,
	page int,
	limit int,
) (*ProductReviewsResponse, error) {

	if sort == "" {
		sort = SortRecent
	}
	orderBy, ok := sortOrders[sort]
	if !ok {
		return nil, ErrInvalidSort
	}

	exists, err := s.repository.ProductExists(ctx, productID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrProductNotFound
	}

	reviews, total, err := s.repository.ListByProduct(ctx, productID, rating, orderBy, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}

	histogram, err := s.repository.RatingHistogram(ctx, productID)
	if err != nil {
		return nil, err
	}

	summary := RatingSummary{Histogram: make(map[string]int64, 5)}
	var sum int64
	for stars := 1; stars <= 5; stars++ {
		count := histogram[stars]
		summary.Histogram[strconv.Itoa(stars)] = count
		summary.Count += count
		sum += int64(stars) * count
	}
	if summary.Count > 0 {
		summary.Average = float64(sum*100/summary.Count) / 100
	}

	items := make([]ReviewResponse, 0, len(reviews))
	for i := range reviews {
		items = append(items, toResponse(&reviews[i]))
	}

	return &ProductReviewsResponse{Items: items, Total: total, Page: page, Limit: limit, Summary: summary}, nil
}

// MyReview returns the caller's review of a product.
func (s *Service) MyReview(ctx context.Context, principal authz.Principal, productID uuid.UUID) (*ReviewResponse, error) {
	review, err := s.repository.GetByProductAndUser(ctx, productID, principal.UserID)
	if err != nil {
		return nil, err
	}
	if review == nil {
		return nil, ErrReviewNotFound
	}

	response := toResponse(review)
	return &response, nil
}

// UpsertReview creates the caller's review of a product, or replaces it if
// they already reviewed it, and refreshes the product's rating.
func (s *Service) UpsertReview(
	ctx context.Context,
	principal authz.Principal,
	productID uuid.UUID,
	req UpsertReviewRequest,
) (*ReviewResponse, error) {

	exists, err := s.repository.ProductExists(ctx, productID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrProductNotFound
	}

	verified, err := s.repository.HasPurchased(ctx, principal.UserID, productID)
	if err != nil {
		return nil, err
	}

	tx := s.repository.Begin(ctx)
	if tx.Error != nil {
		return nil, tx.Error
	}
	defer rollbackOnPanic(tx)

	if err := s.repository.LockProductTx(tx, productID); err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	review, err := s.repository.GetByProductAndUserTx(tx, productID, principal.UserID)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if review == nil {
		review = &models.Review{ProductID: productID, UserID: principal.UserID}
	}

	review.Rating = req.Rating
	review.Title = strings.TrimSpace(req.Title)
	review.Body = strings.TrimSpace(req.Body)
	review.VerifiedPurchase = verified

	if err := s.repository.SaveTx(tx, review); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := s.repository.RefreshProductRatingTx(tx, productID); err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	logger.InfoContext(ctx, s.log, "review saved",
		zap.String("review_id", review.ID.String()),
		zap.String("product_id", productID.String()),
		zap.Int("rating", review.Rating),
	)

	return s.MyReview(ctx, principal, productID)
}

// DeleteReview removes a review. Customers may delete only their own;
// others are reported as not found.
func (s *Service) DeleteReview(ctx context.Context, principal authz.Principal, id uuid.UUID) error {
	review, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrReviewNotFound
		}
		return err
	}

	if !principal.CanAccess(review.UserID) {
		return ErrReviewNotFound
	}

	tx := s.repository.Begin(ctx)
	if tx.Error != nil {
		return tx.Error
	}
	defer rollbackOnPanic(tx)

	if err := s.repository.LockProductTx(tx, review.ProductID); err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return err
	}

	if err := s.repository.DeleteTx(tx, review); err != nil {
		tx.Rollback()
		return err
	}

	if err := s.repository.RefreshProductRatingTx(tx, review.ProductID); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	logger.InfoContext(ctx, s.log, "review deleted",
		zap.String("review_id", review.ID.String()),
		zap.String("product_id", review.ProductID.String()),
	)

	return nil
}

// ListReviews returns one page of all reviews for moderation. Callers must
// restrict it to admins.
func (s *Service) ListReviews(ctx context.Context, rating, page, limit int) (*AdminReviewListResponse, error) {
	reviews, total, err := s.repository.ListAll(ctx, rating, limit, (page-1)*limit)
	if err != nil {
		return nil, err
	}

	items := make([]AdminReviewResponse, 0, len(reviews))
	for i := range reviews {
		items = append(items, AdminReviewResponse{
			ReviewResponse: toResponse(&reviews[i]),
			ProductName:    reviews[i].Product.Name,
			ReviewerEmail:  reviews[i].User.Email,
		})
	}

	return &AdminReviewListResponse{Items: items, Total: total, Page: page, Limit: limit}, nil
}

func toResponse(review *models.Review) ReviewResponse {
	return ReviewResponse{
		ID:               review.ID,
		ProductID:        review.ProductID,
		UserID:           review.UserID,
		ReviewerName:     reviewerName(review.User),
		Rating:           review.Rating,
		Title:            review.Title,
		Body:             review.Body,
		VerifiedPurchase: review.VerifiedPurchase,
		CreatedAt:        review.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        review.UpdatedAt.Format(time.RFC3339),
	}
}

// reviewerName shows the first name and last initial ("Asha R."); deleted
// accounts show as "Customer".
func reviewerName(user models.User) string {
	first := strings.TrimSpace(user.FirstName)
	if first == "" {
		return "Customer"
	}
	if last := []rune(strings.TrimSpace(user.LastName)); len(last) > 0 {
		return first + " " + strings.ToUpper(string(last[0])) + "."
	}
	return first
}

func rollbackOnPanic(tx *gorm.DB) {
	if r := recover(); r != nil {
		tx.Rollback()
		panic(r)
	}
}
