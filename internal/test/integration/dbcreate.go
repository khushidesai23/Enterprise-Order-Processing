package integration

import (
	"gorm.io/gorm"
)

type gormCreator struct {
	db *gorm.DB
}

func (c gormCreator) Create(value interface{}) error {
	return c.db.Create(value).Error
}

func GormCreator(db *gorm.DB) interface{ Create(value interface{}) error } {
	return gormCreator{db: db}
}
