package models

import "github.com/google/uuid"

type Inventory struct {
	BaseModel

	ProductID uuid.UUID `gorm:"uniqueIndex;not null"`
	Product Product `gorm:"foreignKey:ProductID"`

	AvailableQuantity int `gorm:"default:0"`
	ReservedQuantity  int `gorm:"default:0"`
}