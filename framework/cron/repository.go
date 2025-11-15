package cron

import (
	"context"
	"fmt"
	"github.com/hawthorntrees/cronframework/framework/dbs"
	"github.com/hawthorntrees/cronframework/framework/model"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"time"
)

type Repository struct {
	db func() *gorm.DB
}

func NewRepository() *Repository {
	return &Repository{
		db: dbs.GetDB,
	}
}

func (r *Repository) GetEnabledTasks(ctx context.Context) ([]*model.Hawthorn_task, error) {
	var tasks []*model.Hawthorn_task
	result := r.db().WithContext(ctx).Where("enabled = ?", true).Find(&tasks)
	if result.Error != nil {
		return nil, fmt.Errorf("查询任务失败: %w", result.Error)
	}
	return tasks, nil
}

func (r *Repository) TryLockTask(ctx context.Context, taskID int64, now time.Time, expiredAt time.Time) error {
	err := r.db().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Exec("SET LOCAL statement_timeout = 10000").Error; err != nil {
			return err
		}
		var lockTask model.Hawthorn_task
		result := tx.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE SKIP LOCKED").
			Model(&lockTask).
			Where("id=? and enabled = true and (expired_at is null or expired_at < ?)", taskID, now).
			Updates(map[string]interface{}{
				"locked_at":  now,
				"expired_at": expiredAt,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	return err
}

func (r *Repository) ReleaseLockTask(ctx context.Context, taskID int64, now time.Time, expiredAt time.Time, lg *zap.Logger) error {
	err := r.db().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Exec("SET LOCAL statement_timeout = 10011").Error; err != nil {
			return err
		}

		var lockTask model.Hawthorn_task
		result := tx.WithContext(ctx).Set("gorm:query_option", "FOR UPDATE SKIP LOCKED").
			Model(&lockTask).
			Where("id=? and locked_at=? and expired_at=?", taskID, now, expiredAt).
			Updates(map[string]interface{}{
				"locked_at":  nil,
				"expired_at": nil,
			})
		return result.Error
	})
	return err
}

func (r *Repository) CreateExecution(ctx context.Context, execution *model.Hawthorn_task_execution) error {
	result := r.db().WithContext(ctx).Create(execution)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
