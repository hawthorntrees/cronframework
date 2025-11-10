package dbs

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func InsertOrUpdate(ctx context.Context, db *gorm.DB, data interface{}) *gorm.DB {
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(data)
}

func InsertOrNothing(ctx context.Context, db *gorm.DB, data interface{}) *gorm.DB {
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(data)
}
