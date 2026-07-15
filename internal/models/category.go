package models

type Category struct {
	BaseModel

	Name string `gorm:"size:100;uniqueIndex;not null"`

	Description string `gorm:"size:500"`

	Products []Product
}