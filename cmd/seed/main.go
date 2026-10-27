// Command seed fills the database with a large demo store: the DummyJSON
// catalogue expanded into product variants with images, stock levels,
// customers, a year of orders with payments, and reviews.
//
//	go run ./cmd/seed                      # defaults below
//	go run ./cmd/seed -customers 5000 -orders 100000
//
// It reads the same configuration (.env) as the API and writes directly
// to the database, so payments are recorded without calling Razorpay:
// seeded payments cannot be refunded through the gateway. Re-running is
// safe: existing products and customers are kept, and orders and reviews
// are only generated once unless -force-orders is given.
package main

import (
	"errors"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	gormlogger "gorm.io/gorm/logger"

	"github.com/khushidesai23/Enterprise-Order-Processing/config"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/database"
	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/utils"
)

const customerDomain = "shopper.example.com"

type options struct {
	customers   int
	orders      int
	days        int
	reviewRate  float64
	password    string
	seed        int64
	forceOrders bool
}

func main() {
	var opts options
	flag.IntVar(&opts.customers, "customers", 1500, "customer accounts to create")
	flag.IntVar(&opts.orders, "orders", 25000, "orders to generate")
	flag.IntVar(&opts.days, "days", 365, "days of order history")
	flag.Float64Var(&opts.reviewRate, "review-rate", 0.35, "share of delivered items that get a review")
	flag.StringVar(&opts.password, "password", "password123", "password for every seeded customer")
	flag.Int64Var(&opts.seed, "seed", 42, "random seed (same seed, same data)")
	flag.BoolVar(&opts.forceOrders, "force-orders", false, "generate orders and reviews even if seeded customers already have orders")
	flag.Parse()

	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "seed:", err)
		os.Exit(1)
	}
}

func run(opts options) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	database, err := database.New(cfg)
	if err != nil {
		return err
	}
	defer database.Close()

	if err := database.AutoMigrate(); err != nil {
		return err
	}

	db := database.DB.Session(&gorm.Session{Logger: gormlogger.Default.LogMode(gormlogger.Silent)})
	rng := rand.New(rand.NewSource(opts.seed))
	started := time.Now()

	source, err := loadCatalog()
	if err != nil {
		return err
	}

	categories, err := seedCategories(db)
	if err != nil {
		return err
	}

	listings := buildListings(source)
	if err := seedProducts(db, rng, listings, categories); err != nil {
		return err
	}

	customers, err := seedCustomers(db, rng, opts)
	if err != nil {
		return err
	}

	var seededOrders int64
	if err := db.Model(&models.Order{}).
		Joins("JOIN users ON users.id = orders.user_id").
		Where("users.email LIKE ?", "%@"+customerDomain).
		Count(&seededOrders).Error; err != nil {
		return err
	}

	if seededOrders > 0 && !opts.forceOrders {
		fmt.Printf("orders: %d seeded orders already exist, skipping orders and reviews (use -force-orders to add more)\n", seededOrders)
	} else if opts.orders > 0 {
		delivered, err := seedOrders(db, rng, opts, listings, customers)
		if err != nil {
			return err
		}
		if err := seedReviews(db, rng, opts, listings, customers, delivered); err != nil {
			return err
		}
	}

	if err := refreshRatings(db); err != nil {
		return err
	}

	fmt.Printf("seed: done in %s\n", time.Since(started).Round(time.Second))
	return nil
}

func seedCategories(db *gorm.DB) (map[string]uuid.UUID, error) {
	ids := make(map[string]uuid.UUID, len(departments))
	created := 0

	slugs := make([]string, 0, len(departments))
	for slug := range departments {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)

	for _, slug := range slugs {
		dept := departments[slug]
		var category models.Category
		err := db.Where("name = ?", dept.Name).First(&category).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			category = models.Category{Name: dept.Name, Description: dept.Description}
			if err := db.Omit(clause.Associations).Create(&category).Error; err != nil {
				return nil, fmt.Errorf("create category %s: %w", dept.Name, err)
			}
			created++
		} else if err != nil {
			return nil, err
		}
		ids[slug] = category.ID
	}

	fmt.Printf("categories: %d departments (%d new)\n", len(ids), created)
	return ids, nil
}

func seedProducts(db *gorm.DB, rng *rand.Rand, listings []listing, categories map[string]uuid.UUID) error {
	skus := make([]string, len(listings))
	for i := range listings {
		skus[i] = listings[i].product.SKU
	}

	existing, err := productIDsBySKU(db, skus)
	if err != nil {
		return err
	}

	var fresh []*models.Product
	for i := range listings {
		product := &listings[i].product
		product.CategoryID = categories[listings[i].department]
		if id, ok := existing[product.SKU]; ok {
			product.ID = id
			continue
		}
		fresh = append(fresh, product)
	}

	if len(fresh) > 0 {
		if err := db.Omit(clause.Associations).CreateInBatches(fresh, 200).Error; err != nil {
			return fmt.Errorf("create products: %w", err)
		}

		inventories := make([]models.Inventory, 0, len(fresh))
		for _, product := range fresh {
			inventories = append(inventories, models.Inventory{
				ProductID:         product.ID,
				AvailableQuantity: randomStock(rng),
			})
		}
		if err := db.Omit(clause.Associations).CreateInBatches(inventories, 500).Error; err != nil {
			return fmt.Errorf("create inventory: %w", err)
		}
	}

	fmt.Printf("products: %d listings with images (%d new)\n", len(listings), len(fresh))
	return nil
}

// randomStock gives most products healthy stock, some low and a few none.
func randomStock(rng *rand.Rand) int {
	switch roll := rng.Float64(); {
	case roll < 0.06:
		return 0
	case roll < 0.16:
		return 1 + rng.Intn(9)
	default:
		return 20 + rng.Intn(580)
	}
}

func productIDsBySKU(db *gorm.DB, skus []string) (map[string]uuid.UUID, error) {
	ids := make(map[string]uuid.UUID, len(skus))
	for start := 0; start < len(skus); start += 500 {
		end := min(start+500, len(skus))
		var rows []struct {
			ID  uuid.UUID
			SKU string
		}
		if err := db.Model(&models.Product{}).Select("id, sku").Where("sku IN ?", skus[start:end]).Scan(&rows).Error; err != nil {
			return nil, err
		}
		for _, row := range rows {
			ids[row.SKU] = row.ID
		}
	}
	return ids, nil
}

type customer struct {
	ID        uuid.UUID
	CreatedAt time.Time
}

func seedCustomers(db *gorm.DB, rng *rand.Rand, opts options) ([]customer, error) {
	hash, err := utils.HashPassword(opts.password)
	if err != nil {
		return nil, err
	}

	var existing []models.User
	if err := db.Select("id, email, created_at").Where("email LIKE ?", "%@"+customerDomain).Find(&existing).Error; err != nil {
		return nil, err
	}
	taken := make(map[string]bool, len(existing))
	for _, user := range existing {
		taken[user.Email] = true
	}

	now := time.Now()
	var fresh []models.User
	for i := 1; i <= opts.customers; i++ {
		first := firstNames[(i*7)%len(firstNames)]
		last := lastNames[(i*13)%len(lastNames)]
		email := fmt.Sprintf("%s.%s%d@%s",
			strings.ToLower(first),
			strings.ToLower(strings.NewReplacer("'", "", " ", "").Replace(last)),
			i, customerDomain)
		if taken[email] {
			continue
		}

		joined := now.Add(-time.Duration(rng.Int63n(int64(730 * 24 * time.Hour))))
		fresh = append(fresh, models.User{
			BaseModel: models.BaseModel{CreatedAt: joined, UpdatedAt: joined},
			FirstName: first,
			LastName:  last,
			Email:     email,
			Password:  hash,
			IsActive:  rng.Float64() > 0.02,
			Role:      models.RoleCustomer,
		})
	}

	if len(fresh) > 0 {
		if err := db.CreateInBatches(fresh, 500).Error; err != nil {
			return nil, fmt.Errorf("create customers: %w", err)
		}
	}

	customers := make([]customer, 0, len(existing)+len(fresh))
	for _, user := range append(existing, fresh...) {
		customers = append(customers, customer{ID: user.ID, CreatedAt: user.CreatedAt})
	}

	fmt.Printf("customers: %d (%d new, password %q)\n", len(customers), len(fresh), opts.password)
	return customers, nil
}

// deliveredItem is a delivered purchase that may be reviewed.
type deliveredItem struct {
	userID      uuid.UUID
	listing     int
	deliveredAt time.Time
}

// statusAt picks a realistic order status for an order placed age ago.
func statusAt(rng *rand.Rand, age time.Duration) (models.OrderStatus, bool) {
	roll := rng.Float64()
	days := age.Hours() / 24

	switch {
	case days > 14:
		switch {
		case roll < 0.80:
			return models.OrderDelivered, false
		case roll < 0.82:
			return models.OrderDelivered, true // returned and refunded after delivery
		case roll < 0.86:
			return models.OrderCancelled, true // refunded before shipping
		default:
			return models.OrderCancelled, false // abandoned or cancelled unpaid
		}
	case days > 3:
		switch {
		case roll < 0.45:
			return models.OrderDelivered, false
		case roll < 0.80:
			return models.OrderShipped, false
		case roll < 0.85:
			return models.OrderPacked, false
		case roll < 0.88:
			return models.OrderCancelled, true
		default:
			return models.OrderCancelled, false
		}
	default:
		switch {
		case roll < 0.30:
			return models.OrderPaid, false
		case roll < 0.48:
			return models.OrderPacked, false
		case roll < 0.58:
			return models.OrderShipped, false
		case roll < 0.68:
			return models.OrderCreated, false
		case roll < 0.82:
			return models.OrderPaymentPending, false
		default:
			return models.OrderCancelled, false
		}
	}
}

func seedOrders(db *gorm.DB, rng *rand.Rand, opts options, listings []listing, customers []customer) ([]deliveredItem, error) {
	if len(customers) == 0 {
		return nil, errors.New("no customers to place orders")
	}

	// Cumulative popularity for weighted product picks.
	cumulative := make([]float64, len(listings))
	total := 0.0
	for i := range listings {
		total += listings[i].popularity
		cumulative[i] = total
	}
	pickListing := func() int {
		return sort.SearchFloat64s(cumulative, rng.Float64()*total)
	}

	now := time.Now()
	window := time.Duration(opts.days) * 24 * time.Hour

	orders := make([]models.Order, 0, opts.orders)
	itemsPerOrder := make([][]models.OrderItem, 0, opts.orders)
	payments := make([]*models.Payment, 0, opts.orders)
	reserved := make(map[int]int)
	var delivered []deliveredItem

	for n := 0; n < opts.orders; n++ {
		buyer := customers[rng.Intn(len(customers))]

		// Busier recently: sqrt skews ages towards zero.
		age := time.Duration((1 - math.Sqrt(rng.Float64())) * float64(window))
		placed := now.Add(-age)
		if placed.Before(buyer.CreatedAt) {
			placed = buyer.CreatedAt.Add(time.Duration(rng.Int63n(int64(48 * time.Hour))))
			if placed.After(now) {
				placed = now.Add(-time.Minute)
			}
			age = now.Sub(placed)
		}

		status, refunded := statusAt(rng, age)

		lines := 1
		switch roll := rng.Float64(); {
		case roll > 0.93:
			lines = 4
		case roll > 0.82:
			lines = 3
		case roll > 0.55:
			lines = 2
		}

		seen := map[int]bool{}
		var items []models.OrderItem
		var amount float64
		var picked []int
		for len(items) < lines {
			index := pickListing()
			if seen[index] {
				if len(seen) >= len(listings) {
					break
				}
				continue
			}
			seen[index] = true

			product := &listings[index].product
			quantity := 1
			if product.Price < 2000 && rng.Float64() < 0.25 {
				quantity = 2 + rng.Intn(2)
			}

			items = append(items, models.OrderItem{
				BaseModel: models.BaseModel{CreatedAt: placed, UpdatedAt: placed},
				ProductID: product.ID,
				Quantity:  quantity,
				Price:     product.Price,
			})
			picked = append(picked, index)
			amount += product.Price * float64(quantity)
		}

		updated := placed.Add(time.Duration(rng.Int63n(int64(min(age, 72*time.Hour)) + 1)))
		orders = append(orders, models.Order{
			BaseModel:   models.BaseModel{CreatedAt: placed, UpdatedAt: updated},
			UserID:      buyer.ID,
			Status:      status,
			TotalAmount: math.Round(amount*100) / 100,
		})
		itemsPerOrder = append(itemsPerOrder, items)
		payments = append(payments, paymentFor(rng, status, refunded, amount, placed))

		switch status {
		case models.OrderCreated, models.OrderPaymentPending, models.OrderPaid, models.OrderPacked:
			for i, index := range picked {
				reserved[index] += items[i].Quantity
			}
		case models.OrderDelivered:
			deliveredAt := placed.Add(time.Duration(2+rng.Intn(5)) * 24 * time.Hour)
			if deliveredAt.After(now) {
				deliveredAt = now
			}
			for _, index := range picked {
				delivered = append(delivered, deliveredItem{userID: buyer.ID, listing: index, deliveredAt: deliveredAt})
			}
		}
	}

	return delivered, db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit(clause.Associations).CreateInBatches(orders, 1000).Error; err != nil {
			return fmt.Errorf("create orders: %w", err)
		}

		var allItems []models.OrderItem
		var allPayments []models.Payment
		for i := range orders {
			for _, item := range itemsPerOrder[i] {
				item.OrderID = orders[i].ID
				allItems = append(allItems, item)
			}
			if payment := payments[i]; payment != nil {
				payment.OrderID = orders[i].ID
				allPayments = append(allPayments, *payment)
			}
		}

		if err := tx.Omit(clause.Associations).CreateInBatches(allItems, 2000).Error; err != nil {
			return fmt.Errorf("create order items: %w", err)
		}
		if err := tx.Omit(clause.Associations).CreateInBatches(allPayments, 2000).Error; err != nil {
			return fmt.Errorf("create payments: %w", err)
		}

		// Open orders hold their stock, as they would in the live system.
		for index, quantity := range reserved {
			if err := tx.Model(&models.Inventory{}).
				Where("product_id = ?", listings[index].product.ID).
				Update("reserved_quantity", gorm.Expr("reserved_quantity + ?", quantity)).Error; err != nil {
				return fmt.Errorf("reserve stock: %w", err)
			}
		}

		counts := map[models.OrderStatus]int{}
		for _, order := range orders {
			counts[order.Status]++
		}
		fmt.Printf("orders: %d over %d days, %d items, %d payments %v\n",
			len(orders), opts.days, len(allItems), len(allPayments), counts)
		return nil
	})
}

// paymentFor records the payment an order in this state would have.
func paymentFor(rng *rand.Rand, status models.OrderStatus, refunded bool, amount float64, placed time.Time) *models.Payment {
	var paymentStatus models.PaymentStatus
	switch {
	case refunded:
		paymentStatus = models.PaymentRefunded
	case status == models.OrderCreated:
		return nil
	case status == models.OrderPaymentPending:
		paymentStatus = models.PaymentPending
		if rng.Float64() < 0.3 {
			paymentStatus = models.PaymentFailed
		}
	case status == models.OrderCancelled:
		// Cancelled before paying: most never started checkout.
		switch roll := rng.Float64(); {
		case roll < 0.6:
			return nil
		case roll < 0.85:
			paymentStatus = models.PaymentFailed
		default:
			paymentStatus = models.PaymentPending
		}
	default:
		paymentStatus = models.PaymentSuccess
	}

	created := placed.Add(time.Duration(1+rng.Intn(10)) * time.Minute)
	payment := &models.Payment{
		BaseModel:      models.BaseModel{CreatedAt: created, UpdatedAt: created},
		GatewayOrderID: "order_seed_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:16],
		Status:         paymentStatus,
		Amount:         math.Round(amount*100) / 100,
		Currency:       "INR",
		Gateway:        "RAZORPAY",
	}
	if paymentStatus != models.PaymentPending {
		transaction := "pay_seed_" + strings.ReplaceAll(uuid.NewString(), "-", "")[:16]
		payment.TransactionID = &transaction
	}
	return payment
}

func seedReviews(db *gorm.DB, rng *rand.Rand, opts options, listings []listing, customers []customer, delivered []deliveredItem) error {
	now := time.Now()
	type key struct {
		user    uuid.UUID
		product uuid.UUID
	}
	written := map[key]bool{}
	var reviews []models.Review

	sampleRating := func(base float64) int {
		rating := int(math.Round(base + 0.5 + rng.NormFloat64()*1.1))
		return max(1, min(5, rating))
	}

	for _, item := range delivered {
		if rng.Float64() >= opts.reviewRate {
			continue
		}
		l := &listings[item.listing]
		k := key{item.userID, l.product.ID}
		if written[k] {
			continue
		}
		written[k] = true

		rating := sampleRating(l.baseRating)
		title, body := reviewText(rng, rating, productNoun[l.department])
		at := item.deliveredAt.Add(time.Duration(1+rng.Intn(20)) * 24 * time.Hour)
		if at.After(now) {
			at = now.Add(-time.Duration(rng.Intn(3600)) * time.Second)
		}

		reviews = append(reviews, models.Review{
			BaseModel:        models.BaseModel{CreatedAt: at, UpdatedAt: at},
			ProductID:        l.product.ID,
			UserID:           item.userID,
			Rating:           rating,
			Title:            title,
			Body:             body,
			VerifiedPurchase: true,
		})
	}

	// DummyJSON's own short reviews go on each base listing.
	for i := range listings {
		l := &listings[i]
		if !l.isBase {
			continue
		}
		for _, source := range l.source.Reviews {
			buyer := customers[rng.Intn(len(customers))]
			k := key{buyer.ID, l.product.ID}
			if written[k] || source.Rating < 1 || source.Rating > 5 {
				continue
			}
			written[k] = true

			title, _ := reviewText(rng, source.Rating, productNoun[l.department])
			at := now.Add(-time.Duration(rng.Int63n(int64(time.Duration(opts.days) * 24 * time.Hour))))
			reviews = append(reviews, models.Review{
				BaseModel: models.BaseModel{CreatedAt: at, UpdatedAt: at},
				ProductID: l.product.ID,
				UserID:    buyer.ID,
				Rating:    source.Rating,
				Title:     title,
				Body:      source.Comment,
			})
		}
	}

	if len(reviews) > 0 {
		// Skip any (product, user) pair reviewed in an earlier run.
		if err := db.Clauses(clause.OnConflict{DoNothing: true}).Omit(clause.Associations).CreateInBatches(reviews, 1000).Error; err != nil {
			return fmt.Errorf("create reviews: %w", err)
		}
	}

	fmt.Printf("reviews: %d\n", len(reviews))
	return nil
}

// refreshRatings recomputes every product's rating from its reviews.
func refreshRatings(db *gorm.DB) error {
	err := db.Exec(`
		UPDATE products p SET rating_average = s.average, rating_count = s.count
		FROM (
			SELECT product_id, ROUND(AVG(rating)::numeric, 2) AS average, COUNT(*) AS count
			FROM reviews WHERE deleted_at IS NULL GROUP BY product_id
		) s
		WHERE p.id = s.product_id`).Error
	if err != nil {
		return fmt.Errorf("refresh ratings: %w", err)
	}
	fmt.Println("ratings: refreshed")
	return nil
}
