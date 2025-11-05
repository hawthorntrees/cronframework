package cron

import (
	"context"
	"fmt"
	"github.com/hawthorntrees/cronframework/framework/dbs"
	"github.com/hawthorntrees/cronframework/framework/model"
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

func (r *Repository) TryLockTask(ctx context.Context, taskID int64, timeout time.Duration) error {
	err := r.db().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SET LOCAL statement_timeout = 10000").Error; err != nil {
			return err
		}
		now := time.Now()
		expiredAt := now.Add(timeout)

		var lockTask model.Hawthorn_task
		result := tx.Set("gorm:query_option", "FOR UPDATE SKIP LOCKED").
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

func (r *Repository) CreateExecution(ctx context.Context, execution *model.Hawthorn_task_execution) error {
	result := r.db().WithContext(ctx).Create(execution)
	if result.Error != nil {
		return result.Error
	}
	return nil
}
