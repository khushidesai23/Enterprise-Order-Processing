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

func CreateOrderFixture(
	t testing.TB,
	db interface{ Create(value interface{}) error },
	userID uuid.UUID,
	status models.OrderStatus,
	total float64,
) *models.Order {
	t.Helper()

	order := &models.Order{
		UserID:      userID,
		Status:      status,
		TotalAmount: total,
	}

	if err := db.Create(order); err != nil {
		t.Fatalf("creating order fixture failed: %v", err)
	}

	return order
}

func CreateOrderItemFixture(
	t testing.TB,
	db interface{ Create(value interface{}) error },
	orderID uuid.UUID,
	productID uuid.UUID,
	quantity int,
	price float64,
) *models.OrderItem {
	t.Helper()

	item := &models.OrderItem{
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
		Price:     price,
	}

	if err := db.Create(item); err != nil {
		t.Fatalf("creating order item fixture failed: %v", err)
	}

	return item
}

func CreatePaymentFixture(
	t testing.TB,
	db interface{ Create(value interface{}) error },
	orderID uuid.UUID,
	gatewayOrderID string,
	status models.PaymentStatus,
	amount float64,
) *models.Payment {
	t.Helper()

	payment := &models.Payment{
		OrderID:        orderID,
		GatewayOrderID: gatewayOrderID,
		Status:         status,
		Amount:         amount,
		Currency:       "INR",
		Gateway:        "RAZORPAY",
	}

	if err := db.Create(payment); err != nil {
		t.Fatalf("creating payment fixture failed: %v", err)
	}

	return payment
}
