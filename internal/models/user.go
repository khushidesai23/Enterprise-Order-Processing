package models

type User struct {
	BaseModel

	FirstName string `gorm:"size:100;not null" json:"first_name"`
	LastName  string `gorm:"size:100" json:"last_name"`

	Email string `gorm:"size:255;uniqueIndex;not null" json:"email"`

	Password string `gorm:"size:255;not null" json:"-"`

	IsActive bool `gorm:"default:true" json:"is_active"`
}