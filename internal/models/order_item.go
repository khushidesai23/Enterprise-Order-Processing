package models

import "github.com/google/uuid"

type OrderItem struct {
	BaseModel

	OrderID uuid.UUID

	ProductID uuid.UUID

	Order Order

	Product Product

	Quantity int

	Price float64 `gorm:"type:numeric(12,2)"`
}
