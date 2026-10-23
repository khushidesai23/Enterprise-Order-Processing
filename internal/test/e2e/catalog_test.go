//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	testintegration "github.com/khushidesai23/Enterprise-Order-Processing/internal/test/integration"
)

type searchResult struct {
	Items []struct {
		ID            uuid.UUID `json:"id"`
		Name          string    `json:"name"`
		Brand         string    `json:"brand"`
		Price         float64   `json:"price"`
		MRP           float64   `json:"mrp"`
		ImageURL      string    `json:"image_url"`
		RatingAverage float64   `json:"rating_average"`
		RatingCount   int       `json:"rating_count"`
	} `json:"items"`
	Total  int64 `json:"total"`
	Facets struct {
		Categories []struct {
			Name  string `json:"name"`
			Count int64  `json:"count"`
		} `json:"categories"`
		Brands []struct {
			Name  string `json:"name"`
			Count int64  `json:"count"`
		} `json:"brands"`
		PriceMin float64 `json:"price_min"`
		PriceMax float64 `json:"price_max"`
	} `json:"facets"`
}

func (l *lifecycle) createCatalogProduct(categoryID uuid.UUID, name, sku, brand string, price, mrp float64, stock int) uuid.UUID {
	l.t.Helper()

	code, api := l.call(http.MethodPost, "/api/v1/products", map[string]interface{}{
		"name": name, "sku": sku, "brand": brand, "price": price, "mrp": mrp, "category_id": categoryID,
		"image_url": "https://example.com/" + sku + ".jpg",
		"images":    []string{"https://example.com/" + sku + ".jpg", "https://example.com/" + sku + "-2.jpg"},
		"specs":     []map[string]string{{"name": "Warranty", "value": "1 year"}},
	}, l.adminToken)
	require.Equal(l.t, http.StatusCreated, code, api.Error)

	var product struct {
		ID uuid.UUID `json:"id"`
	}
	testintegration.DecodeData(l.t, api.Data, &product)

	code, api = l.call(http.MethodPost, "/api/v1/inventory",
		map[string]interface{}{"product_id": product.ID, "available_quantity": stock}, l.adminToken)
	require.Equal(l.t, http.StatusCreated, code, api.Error)

	return product.ID
}

func (l *lifecycle) search(query url.Values) searchResult {
	l.t.Helper()

	code, api := l.call(http.MethodGet, "/api/v1/products/search?"+query.Encode(), nil, "")
	require.Equal(l.t, http.StatusOK, code, api.Error)

	var result searchResult
	testintegration.DecodeData(l.t, api.Data, &result)

	return result
}

func TestProductSearchFiltersSortsAndFacets(t *testing.T) {
	l := newLifecycle(t)
	creator := testintegration.GormCreator(l.env.DB.DB)
	audio := testintegration.CreateCategoryFixture(t, creator, "Audio")

	headphones := l.createCatalogProduct(audio.ID, "Wireless Headphones", "AUD-1", "Sonic", 2999, 4999, 10)
	l.createCatalogProduct(audio.ID, "Wired Headphones", "AUD-2", "Sonic", 999, 0, 0)
	l.createCatalogProduct(audio.ID, "Bluetooth Speaker", "AUD-3", "Boomly", 1999, 2499, 5)

	t.Run("words must all match; facets describe the result", func(t *testing.T) {
		result := l.search(url.Values{"q": {"headphones"}})
		require.EqualValues(t, 2, result.Total)
		require.Len(t, result.Facets.Brands, 1)
		assert.Equal(t, "Sonic", result.Facets.Brands[0].Name)
		assert.Equal(t, 999.0, result.Facets.PriceMin)
		assert.Equal(t, 2999.0, result.Facets.PriceMax)

		result = l.search(url.Values{"q": {"wireless sonic"}})
		require.EqualValues(t, 1, result.Total)
		assert.Equal(t, headphones, result.Items[0].ID)
		assert.Equal(t, 4999.0, result.Items[0].MRP)
		assert.Equal(t, "https://example.com/AUD-1.jpg", result.Items[0].ImageURL)
	})

	t.Run("filters", func(t *testing.T) {
		assert.EqualValues(t, 2, l.search(url.Values{"category_id": {audio.ID.String()}, "in_stock": {"true"}}).Total)
		assert.EqualValues(t, 2, l.search(url.Values{"category_id": {audio.ID.String()}, "deals": {"true"}}).Total)
		assert.EqualValues(t, 1, l.search(url.Values{"category_id": {audio.ID.String()}, "brand": {"Boomly"}}).Total)
		assert.EqualValues(t, 3, l.search(url.Values{"category_id": {audio.ID.String()}, "brand": {"Boomly", "Sonic"}}).Total)
		assert.EqualValues(t, 1, l.search(url.Values{"category_id": {audio.ID.String()}, "min_price": {"1500"}, "max_price": {"2500"}}).Total)

		// The brand facet ignores the brand filter, so both brands stay listed.
		assert.Len(t, l.search(url.Values{"category_id": {audio.ID.String()}, "brand": {"Boomly"}}).Facets.Brands, 2)
	})

	t.Run("sorting and paging", func(t *testing.T) {
		result := l.search(url.Values{"category_id": {audio.ID.String()}, "sort": {"price_asc"}, "limit": {"2"}})
		require.Len(t, result.Items, 2)
		assert.EqualValues(t, 3, result.Total)
		assert.Equal(t, 999.0, result.Items[0].Price)
		assert.Equal(t, 1999.0, result.Items[1].Price)

		result = l.search(url.Values{"category_id": {audio.ID.String()}, "sort": {"discount"}})
		assert.Equal(t, headphones, result.Items[0].ID, "40% off beats 20% off")

		l.search(url.Values{"q": {"head"}, "sort": {"bestselling"}})
	})

	t.Run("search input is matched literally", func(t *testing.T) {
		assert.EqualValues(t, 0, l.search(url.Values{"q": {"%"}}).Total)
		assert.EqualValues(t, 0, l.search(url.Values{"q": {"'; DROP TABLE products; --"}}).Total)
	})

	t.Run("invalid parameters", func(t *testing.T) {
		for _, query := range []string{"sort=bogus", "category_id=nope", "min_price=-1", "limit=500"} {
			code, _ := l.call(http.MethodGet, "/api/v1/products/search?"+query, nil, "")
			assert.Equal(t, http.StatusBadRequest, code, query)
		}
	})

	t.Run("MRP below price is rejected", func(t *testing.T) {
		code, _ := l.call(http.MethodPost, "/api/v1/products", map[string]interface{}{
			"name": "Bad", "sku": "AUD-BAD", "price": 100, "mrp": 50, "category_id": audio.ID,
		}, l.adminToken)
		assert.Equal(t, http.StatusBadRequest, code)
	})

	t.Run("updates without catalogue fields keep them", func(t *testing.T) {
		code, api := l.call(http.MethodPut, fmt.Sprintf("/api/v1/products/%s", headphones), map[string]interface{}{
			"name": "Wireless Headphones II", "sku": "AUD-1", "price": 2899, "category_id": audio.ID,
		}, l.adminToken)
		require.Equal(t, http.StatusOK, code, api.Error)

		var product struct {
			Brand  string   `json:"brand"`
			MRP    float64  `json:"mrp"`
			Images []string `json:"images"`
		}
		testintegration.DecodeData(t, api.Data, &product)
		assert.Equal(t, "Sonic", product.Brand)
		assert.Equal(t, 4999.0, product.MRP)
		assert.Len(t, product.Images, 2)
	})
}

func TestProductReviews(t *testing.T) {
	l := newLifecycle(t)
	testintegration.CreateUserFixture(t, testintegration.GormCreator(l.env.DB.DB), "Rev", "Iewer", "reviewer@example.com", "password123")
	reviewerToken := l.login("reviewer@example.com")

	reviews := fmt.Sprintf("/api/v1/products/%s/reviews", l.productID)

	rating := func() (float64, int) {
		t.Helper()
		code, api := l.call(http.MethodGet, fmt.Sprintf("/api/v1/products/%s", l.productID), nil, "")
		require.Equal(t, http.StatusOK, code, api.Error)
		var product struct {
			RatingAverage float64 `json:"rating_average"`
			RatingCount   int     `json:"rating_count"`
		}
		testintegration.DecodeData(t, api.Data, &product)
		return product.RatingAverage, product.RatingCount
	}

	// The customer has bought the product; the reviewer has not.
	orderID := l.createOrder()
	checkout := l.checkout(orderID)
	l.webhook("evt_review_capture", "payment.captured", checkout.GatewayOrderID, "pay_review")
	l.assertState(orderID, models.OrderPaid, checkout.PaymentID, models.PaymentSuccess, remaining, lifecycleQuantity)

	code, _ := l.call(http.MethodPut, reviews, map[string]interface{}{"rating": 5}, "")
	assert.Equal(t, http.StatusUnauthorized, code, "writing needs an account")

	code, api := l.call(http.MethodPut, reviews, map[string]interface{}{"rating": 5, "title": "Great", "body": "Love it"}, l.customerToken)
	require.Equal(t, http.StatusOK, code, api.Error)
	var mine struct {
		ID               uuid.UUID `json:"id"`
		VerifiedPurchase bool      `json:"verified_purchase"`
		ReviewerName     string    `json:"reviewer_name"`
	}
	testintegration.DecodeData(t, api.Data, &mine)
	assert.True(t, mine.VerifiedPurchase)
	assert.Equal(t, "Life C.", mine.ReviewerName)

	code, api = l.call(http.MethodPut, reviews, map[string]interface{}{"rating": 2, "title": "Meh"}, reviewerToken)
	require.Equal(t, http.StatusOK, code, api.Error)
	var theirs struct {
		ID               uuid.UUID `json:"id"`
		VerifiedPurchase bool      `json:"verified_purchase"`
	}
	testintegration.DecodeData(t, api.Data, &theirs)
	assert.False(t, theirs.VerifiedPurchase)

	average, count := rating()
	assert.Equal(t, 3.5, average)
	assert.Equal(t, 2, count)

	// Writing again replaces the review instead of adding one.
	code, api = l.call(http.MethodPut, reviews, map[string]interface{}{"rating": 4}, l.customerToken)
	require.Equal(t, http.StatusOK, code, api.Error)
	average, count = rating()
	assert.Equal(t, 3.0, average)
	assert.Equal(t, 2, count)

	code, api = l.call(http.MethodGet, reviews+"?sort=highest", nil, "")
	require.Equal(t, http.StatusOK, code, api.Error)
	var list struct {
		Items []struct {
			Rating int `json:"rating"`
		} `json:"items"`
		Total   int64 `json:"total"`
		Summary struct {
			Average   float64          `json:"average"`
			Count     int64            `json:"count"`
			Histogram map[string]int64 `json:"histogram"`
		} `json:"summary"`
	}
	testintegration.DecodeData(t, api.Data, &list)
	require.Len(t, list.Items, 2)
	assert.Equal(t, 4, list.Items[0].Rating)
	assert.EqualValues(t, 1, list.Summary.Histogram["2"])
	assert.EqualValues(t, 0, list.Summary.Histogram["5"])

	code, api = l.call(http.MethodGet, reviews+"?rating=2", nil, "")
	require.Equal(t, http.StatusOK, code, api.Error)
	testintegration.DecodeData(t, api.Data, &list)
	assert.EqualValues(t, 1, list.Total)
	assert.EqualValues(t, 2, list.Summary.Count, "the summary covers every review")

	// Only the author or an admin may delete; others get 404.
	code, _ = l.call(http.MethodDelete, fmt.Sprintf("/api/v1/reviews/%s", theirs.ID), nil, l.customerToken)
	assert.Equal(t, http.StatusNotFound, code)
	code, _ = l.call(http.MethodGet, "/api/v1/reviews", nil, l.customerToken)
	assert.Equal(t, http.StatusForbidden, code)
	code, api = l.call(http.MethodGet, "/api/v1/reviews", nil, l.adminToken)
	require.Equal(t, http.StatusOK, code, api.Error)

	code, api = l.call(http.MethodDelete, fmt.Sprintf("/api/v1/reviews/%s", theirs.ID), nil, l.adminToken)
	require.Equal(t, http.StatusOK, code, api.Error)
	average, count = rating()
	assert.Equal(t, 4.0, average)
	assert.Equal(t, 1, count)

	code, _ = l.call(http.MethodGet, reviews+"/me", nil, reviewerToken)
	assert.Equal(t, http.StatusNotFound, code)

	// The reviewer can review again after their review was removed.
	code, api = l.call(http.MethodPut, reviews, map[string]interface{}{"rating": 1}, reviewerToken)
	require.Equal(t, http.StatusOK, code, api.Error)

	code, _ = l.call(http.MethodPut, reviews, map[string]interface{}{"rating": 6}, reviewerToken)
	assert.Equal(t, http.StatusBadRequest, code)
	code, _ = l.call(http.MethodGet, fmt.Sprintf("/api/v1/products/%s/reviews", uuid.New()), nil, "")
	assert.Equal(t, http.StatusNotFound, code)
}
