package database

import (
	"context"
	"time"
)

func (d *Database) Healthy() bool {

	sqlDB, err := d.DB.DB()

	if err != nil {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

	defer cancel()

	return sqlDB.PingContext(ctx) == nil
}
