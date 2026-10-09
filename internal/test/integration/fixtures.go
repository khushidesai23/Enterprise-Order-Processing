package integration

import (
	"testing"

	"github.com/google/uuid"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
	"github.com/khushidesai23/Enterprise-Order-Processing/pkg/utils"
)

func CreateUserFixture(
	t testing.TB,
	db interface{ Create(value interface{}) error },
	firstName string,
	lastName string,
	email string,
	password string,
) *models.User {
	t.Helper()

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	user := &models.User{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Password:  hashedPassword,
		IsActive:  true,
	}

	if err := db.Create(user); err != nil {
		t.Fatalf("creating user fixture failed: %v", err)
	}

	return user
}

func CreateCategoryFixture(
	t testing.TB,
	db interface{ Create(value interface{}) error },
	name string,
) *models.Category {
	t.Helper()

	category := &models.Category{
		Name:        name,
		Description: name + " description",
	}

	if err := db.Create(category); err != nil {
		t.Fatalf("creating category fixture failed: %v", err)
	}

	return category
}

func CreateProductFixture(
	t testing.TB,
	db interface{ Create(value interface{}) error },
	categoryID uuid.UUID,
	name string,
	sku string,
	price float64,
) *models.Product {
	t.Helper()

	product := &models.Product{
		Name:        name,
		Description: name + " description",
		SKU:         sku,
		Price:       price,
		CategoryID:  categoryID,
	}

	if err := db.Create(product); err != nil {
		t.Fatalf("creating product fixture failed: %v", err)
	}

	return product
}

func CreateInventoryFixture(
	t testing.TB,
	db interface{ Create(value interface{}) error },
	productID uuid.UUID,
	available int,
	reserved int,
) *models.Inventory {
	t.Helper()

	inventory := &models.Inventory{
		ProductID:         productID,
		AvailableQuantity: available,
		ReservedQuantity:  reserved,
	}

	if err := db.Create(inventory); err != nil {
		t.Fatalf("creating inventory fixture failed: %v", err)
	}

	return inventory
}
