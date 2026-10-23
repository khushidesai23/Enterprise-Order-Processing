package order

import (
	"github.com/google/uuid"

	"github.com/khushidesai23/Enterprise-Order-Processing/internal/models"
)

const (
	DefaultOrderPageSize = 50
	MaxOrderPageSize     = 100
)

type CreateOrderItemRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,gt=0"`
}

type CreateOrderRequest struct {
	UserID uuid.UUID                `json:"-"`
	Items  []CreateOrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

type UpdateOrderStatusRequest struct {
	Status models.OrderStatus `json:"status" binding:"required"`
}

type OrderItemResponse struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	SKU         string    `json:"sku"`

	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`

	SubTotal float64 `json:"sub_total"`
}

type OrderResponse struct {
	ID uuid.UUID `json:"id"`

	UserID uuid.UUID `json:"user_id"`

	Status models.OrderStatus `json:"status"`

	TotalAmount float64 `json:"total_amount"`

	Items []OrderItemResponse `json:"items"`

	CreatedAt string `json:"created_at"`
}

type OrderListResponse struct {
	ID uuid.UUID `json:"id"`

	UserID uuid.UUID `json:"user_id"`

	Status models.OrderStatus `json:"status"`

	TotalAmount float64 `json:"total_amount"`

	TotalItems int `json:"total_items"`

	CreatedAt string `json:"created_at"`
}

// OrderStatusSummary counts orders per status; every status is present.
type OrderStatusSummary struct {
	TotalOrders int64                        `json:"total_orders"`
	ByStatus    map[models.OrderStatus]int64 `json:"by_status"`
}

var orderStatuses = []models.OrderStatus{
	models.OrderCreated,
	models.OrderPaymentPending,
	models.OrderPaid,
	models.OrderPacked,
	models.OrderShipped,
	models.OrderDelivered,
	models.OrderCancelled,
}

// IsValidOrderStatus reports whether status is one of the known statuses.
func IsValidOrderStatus(status models.OrderStatus) bool {
	for _, known := range orderStatuses {
		if status == known {
			return true
		}
	}

	return false
}

type OrderSummary struct {
	TotalOrders      int64   `json:"total_orders"`
	CompletedOrders  int64   `json:"completed_orders"`
	CancelledOrders  int64   `json:"cancelled_orders"`
	RevenueGenerated float64 `json:"revenue_generated"`
}
