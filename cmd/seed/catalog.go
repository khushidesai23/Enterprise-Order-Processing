package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

// data/products.json is the DummyJSON (https://dummyjson.com) product
// catalogue, trimmed to the fields used here. Images stay on DummyJSON's
// CDN.
//
//go:embed data/products.json
var catalogJSON []byte

type sourceProduct struct {
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Discount    float64 `json:"discount"`
	Rating      float64 `json:"rating"`
	Brand       string  `json:"brand"`
	SKU         string  `json:"sku"`
	Weight      float64 `json:"weight"`
	Dimensions  struct {
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
		Depth  float64 `json:"depth"`
	} `json:"dimensions"`
	Warranty  string   `json:"warranty"`
	Shipping  string   `json:"shipping"`
	Returns   string   `json:"returns"`
	Tags      []string `json:"tags"`
	Thumbnail string   `json:"thumbnail"`
	Images    []string `json:"images"`
	Reviews   []struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
		Name    string `json:"name"`
	} `json:"reviews"`
}

func loadCatalog() ([]sourceProduct, error) {
	var products []sourceProduct
	if err := json.Unmarshal(catalogJSON, &products); err != nil {
		return nil, fmt.Errorf("parse catalogue: %w", err)
	}
	return products, nil
}

type department struct {
	Name        string
	Description string
	HouseBrand  string
	// Variant axes: every combination becomes one listing, the first is
	// the source product itself.
	Axes []variantAxis
}

type variantAxis struct {
	Name    string
	Options []variantOption
}

type variantOption struct {
	Label string
	// Price multiplier relative to the first option.
	Factor float64
}

var (
	storage       = variantAxis{Name: "Storage", Options: []variantOption{{"128 GB", 1}, {"256 GB", 1.14}, {"512 GB", 1.32}}}
	phoneColour   = variantAxis{Name: "Colour", Options: []variantOption{{"Midnight Black", 1}, {"Glacier Blue", 1.02}}}
	apparelColour = variantAxis{Name: "Colour", Options: []variantOption{
		{"Classic", 1}, {"Navy Blue", 1}, {"Olive Green", 1.03}, {"Maroon", 1.03}, {"Charcoal Grey", 1.05},
	}}
	shoeColour = variantAxis{Name: "Colour", Options: []variantOption{{"Classic", 1}, {"Black", 1}, {"White", 1.04}, {"Tan", 1.06}}}
	finish     = variantAxis{Name: "Finish", Options: []variantOption{{"Standard", 1}, {"Matte Black", 1.05}, {"Rose Gold", 1.08}, {"Silver", 1.04}}}
	homeColour = variantAxis{Name: "Colour", Options: []variantOption{{"Standard", 1}, {"Ivory", 1.03}, {"Walnut", 1.08}, {"Slate Grey", 1.05}}}
	gearColour = variantAxis{Name: "Colour", Options: []variantOption{{"Standard", 1}, {"Black", 1}, {"Red", 1.02}, {"Blue", 1.02}}}
	beautySize = variantAxis{Name: "Size", Options: []variantOption{{"Standard", 1}, {"Travel Size", 0.55}, {"Value Size", 1.6}}}
	pack       = variantAxis{Name: "Pack", Options: []variantOption{{"Single", 1}, {"Pack of 2", 1.85}}}
	grocery    = variantAxis{Name: "Pack", Options: []variantOption{{"Standard", 1}, {"Pack of 2", 1.9}, {"Pack of 4", 3.6}, {"Family Pack", 5}}}
	vehicle    = variantAxis{Name: "Colour", Options: []variantOption{{"Standard", 1}, {"Pearl White", 1.02}, {"Racing Red", 1.03}}}
)

var departments = map[string]department{
	"beauty":              {"Beauty", "Makeup, cosmetics and beauty essentials.", "Meridian Beauty", []variantAxis{beautySize}},
	"fragrances":          {"Fragrances", "Perfumes and colognes for every occasion.", "Meridian Scents", []variantAxis{beautySize, pack}},
	"furniture":           {"Furniture", "Beds, sofas, tables and storage for every room.", "Meridian Living", []variantAxis{homeColour}},
	"groceries":           {"Groceries", "Fresh produce, pantry staples and household supplies.", "Meridian Fresh", []variantAxis{grocery}},
	"home-decoration":     {"Home Decor", "Lamps, plants, frames and decorative accents.", "Meridian Living", []variantAxis{homeColour}},
	"kitchen-accessories": {"Kitchen & Dining", "Cookware, utensils and kitchen tools.", "Meridian Kitchen", []variantAxis{homeColour}},
	"laptops":             {"Laptops", "Laptops and notebooks for work, study and play.", "Meridian Tech", []variantAxis{storage, phoneColour}},
	"mens-shirts":         {"Men's Clothing", "Shirts and everyday wear for men.", "Meridian Basics", []variantAxis{apparelColour}},
	"mens-shoes":          {"Men's Shoes", "Sneakers, formal and sports shoes for men.", "Meridian Basics", []variantAxis{shoeColour}},
	"mens-watches":        {"Men's Watches", "Analogue, chronograph and smart watches for men.", "Meridian Time", []variantAxis{finish}},
	"mobile-accessories":  {"Mobile Accessories", "Chargers, cases, earphones and more.", "Meridian Tech", []variantAxis{gearColour}},
	"motorcycle":          {"Motorcycles", "Motorcycles and two-wheelers.", "Meridian Motors", []variantAxis{vehicle}},
	"skin-care":           {"Skin Care", "Cleansers, moisturisers and serums.", "Meridian Beauty", []variantAxis{beautySize, pack}},
	"smartphones":         {"Smartphones", "The latest smartphones from top brands.", "Meridian Tech", []variantAxis{storage, phoneColour}},
	"sports-accessories":  {"Sports & Fitness", "Equipment and accessories for every sport.", "Meridian Sports", []variantAxis{gearColour}},
	"sunglasses":          {"Sunglasses", "Sunglasses and eyewear.", "Meridian Optics", []variantAxis{finish}},
	"tablets":             {"Tablets", "Tablets for entertainment and productivity.", "Meridian Tech", []variantAxis{storage, phoneColour}},
	"tops":                {"Women's Tops", "Tops, tees and blouses.", "Meridian Basics", []variantAxis{apparelColour}},
	"vehicle":             {"Cars", "Cars and SUVs.", "Meridian Motors", []variantAxis{vehicle}},
	"womens-bags":         {"Women's Bags", "Handbags, totes and clutches.", "Meridian Atelier", []variantAxis{finish}},
	"womens-dresses":      {"Women's Dresses", "Dresses for every occasion.", "Meridian Atelier", []variantAxis{apparelColour}},
	"womens-jewellery":    {"Jewellery", "Necklaces, earrings and rings.", "Meridian Atelier", []variantAxis{finish}},
	"womens-shoes":        {"Women's Shoes", "Heels, flats and sneakers for women.", "Meridian Atelier", []variantAxis{shoeColour}},
	"womens-watches":      {"Women's Watches", "Elegant and smart watches for women.", "Meridian Time", []variantAxis{finish}},
}

// listing is one product to create, with what the order and review
// generators need to know about it.
type listing struct {
	product    models.Product
	department string
	baseRating float64
	// popularity weights how often the listing is ordered.
	popularity float64
	source     *sourceProduct
	isBase     bool
}

const usdToInr = 83.0

// buildListings expands each source product into its variant listings.
func buildListings(source []sourceProduct) []listing {
	var listings []listing

	for i := range source {
		src := &source[i]
		dept, ok := departments[src.Category]
		if !ok {
			continue
		}

		brand := src.Brand
		if brand == "" {
			brand = dept.HouseBrand
		}

		basePrice := src.Price * usdToInr
		for v, combo := range combinations(dept.Axes) {
			factor := 1.0
			var labels []string
			var specs []models.ProductSpec
			for a, option := range combo {
				factor *= option.Factor
				specs = append(specs, models.ProductSpec{Name: dept.Axes[a].Name, Value: option.Label})
				if option.Label != "Standard" && option.Label != "Classic" && option.Label != "Single" {
					labels = append(labels, option.Label)
				}
			}

			name := src.Title
			if len(labels) > 0 {
				name = fmt.Sprintf("%s (%s)", src.Title, strings.Join(labels, ", "))
			}

			price := charmPrice(basePrice * factor)
			mrp := 0.0
			if src.Discount >= 3 {
				// DummyJSON discounts run 0-20%; widen them a little so the
				// deals pages have range.
				discount := math.Min(src.Discount*1.6, 55)
				mrp = math.Ceil(price/(1-discount/100)/10) * 10
				if mrp <= price {
					mrp = 0
				}
			}

			sku := src.SKU
			if v > 0 {
				sku = fmt.Sprintf("%s-V%d", src.SKU, v)
			}

			specs = append(specs,
				models.ProductSpec{Name: "Brand", Value: brand},
				models.ProductSpec{Name: "Model number", Value: sku},
			)
			if src.Weight > 0 {
				specs = append(specs, models.ProductSpec{Name: "Item weight", Value: fmt.Sprintf("%g kg", src.Weight)})
			}
			if src.Dimensions.Width > 0 {
				specs = append(specs, models.ProductSpec{
					Name:  "Dimensions",
					Value: fmt.Sprintf("%.1f × %.1f × %.1f cm", src.Dimensions.Width, src.Dimensions.Height, src.Dimensions.Depth),
				})
			}
			for _, spec := range []models.ProductSpec{
				{Name: "Warranty", Value: src.Warranty},
				{Name: "Delivery", Value: src.Shipping},
				{Name: "Returns", Value: src.Returns},
			} {
				if spec.Value != "" {
					specs = append(specs, spec)
				}
			}

			images := src.Images
			if len(images) > 10 {
				images = images[:10]
			}
			imageURL := src.Thumbnail
			if len(images) > 0 {
				imageURL = images[0]
			}

			listings = append(listings, listing{
				product: models.Product{
					Name:        truncate(name, 255),
					Description: truncate(src.Description, 1000),
					SKU:         sku,
					Price:       price,
					MRP:         mrp,
					Brand:       brand,
					ImageURL:    imageURL,
					Images:      models.JSONList[string](images),
					Specs:       models.JSONList[models.ProductSpec](specs),
				},
				department: src.Category,
				baseRating: src.Rating,
				// Better-rated, cheaper and default variants sell more:
				// a ₹1,000 item sells ~30x more often than a ₹1,00,000 one.
				popularity: math.Max(src.Rating-1.5, 0.3) / math.Pow(math.Max(price, 100)/1000, 0.75) / float64(v+1),
				source:     src,
				isBase:     v == 0,
			})
		}
	}

	return listings
}

// combinations returns the cartesian product of the axes' options, the
// first combination being every axis's first option.
func combinations(axes []variantAxis) [][]variantOption {
	result := [][]variantOption{{}}
	for _, axis := range axes {
		var next [][]variantOption
		for _, prefix := range result {
			for _, option := range axis.Options {
				combo := append(append([]variantOption{}, prefix...), option)
				next = append(next, combo)
			}
		}
		result = next
	}
	return result
}

// charmPrice rounds to a retail-looking price: ₹49, ₹499, ₹1,299, ₹24,990.
func charmPrice(value float64) float64 {
	switch {
	case value < 100:
		return math.Max(math.Round(value/10)*10-1, 9)
	case value < 10000:
		return math.Round(value/50)*50 - 1
	default:
		return math.Round(value/500)*500 - 10
	}
}

func truncate(value string, max int) string {
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}
