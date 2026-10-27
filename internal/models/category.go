package models

type Category struct {
	BaseModel

	Name string `gorm:"size:100;uniqueIndex:idx_categories_name,where:deleted_at IS NULL;not null"`

	Description string `gorm:"size:500"`

	Products []Product
}
