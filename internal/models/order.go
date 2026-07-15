package models

import "github.com/google/uuid"

type OrderStatus string

const (
	OrderCreated        OrderStatus = "CREATED"
	OrderPaymentPending OrderStatus = "PAYMENT_PENDING"
	OrderPaid           OrderStatus = "PAID"
	OrderPacked         OrderStatus = "PACKED"
	OrderShipped        OrderStatus = "SHIPPED"
	OrderDelivered      OrderStatus = "DELIVERED"
	OrderCancelled      OrderStatus = "CANCELLED"
)

type Order struct {
	BaseModel

	UserID uuid.UUID
	User   User `gorm:"foreignKey:UserID"`

	Status OrderStatus `gorm:"size:30;not null"`

	TotalAmount float64 `gorm:"type:numeric(12,2);not null"`

	Items []OrderItem `gorm:"foreignKey:OrderID"`

	Payment *Payment `gorm:"foreignKey:OrderID"`
}