package models

import "github.com/google/uuid"

type Product struct {
	BaseModel

	Name        string  `gorm:"size:255;not null"`
	Description string  `gorm:"type:text"`
	SKU         string  `gorm:"size:100;uniqueIndex;not null"`
	Price       float64 `gorm:"type:numeric(12,2);not null"`

	CategoryID uuid.UUID
	Category   Category `gorm:"foreignKey:CategoryID"`

	Inventory *Inventory `gorm:"foreignKey:ProductID"`
}