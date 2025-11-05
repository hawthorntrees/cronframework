package model

import (
	"time"
)

type Hawthorn_task struct {
	ID          int64      `gorm:"column:id;type:bigint;primaryKey;autoIncrement" json:"id"`
	Name        string     `gorm:"column:name;type:varchar(100);not null;uniqueIndex" json:"name"`
	Description string     `gorm:"column:description;type:text" json:"description"`
	CronExpr    string     `gorm:"column:cron_expr;type:varchar(100);not null" json:"cron_expr"`
	Enabled     bool       `gorm:"column:enabled;type:bool;not null;default:true" json:"enabled"`
	Timeout     int        `gorm:"column:timeout;type:int;not null;default:300" json:"timeout"` // 秒
	RetryCount  int        `gorm:"column:retry_count;type:int;not null;default:0" json:"retry_count"`
	LockedBy    *string    `gorm:"column:locked_by;type:varchar(100)" json:"locked_by"`
	LockedAt    *time.Time `gorm:"column:locked_at;type:timestamp(3)" json:"locked_at"`
	ExpiredAt   *time.Time `gorm:"column:expired_at;type:timestamp(3)" json:"expired_at"`
	CreatedAt   *time.Time `gorm:"column:created_at;type:timestamp(3)" json:"created_at"`
	UpdatedAt   *time.Time `gorm:"column:updated_at;type:timestamp(3)" json:"updated_at"`
}

func (Hawthorn_task) TableName() string {
	return "hawthorn_task"
}

type Hawthorn_task_execution struct {
	ID          int64      `gorm:"column:id;type:bigint;primaryKey;autoIncrement" json:"id"`
	CreatedDate time.Time  `gorm:"column:created_date;type:date;not null;primaryKey" json:"created_date"`
	TaskID      int64      `gorm:"column:task_id;type:bigint;not null" json:"task_id"`
	NodeID      string     `gorm:"column:node_id;type:varchar(100);not null" json:"node_id"`
	Status      string     `gorm:"column:status;type:varchar(20);not null" json:"status"` // success, failed
	StartTime   time.Time  `gorm:"column:start_time;type:timestamp(3);not null" json:"start_time"`
	EndTime     *time.Time `gorm:"column:end_time;type:timestamp(3)" json:"end_time"`
	Error       string     `gorm:"column:error;type:text" json:"error"`
	TraceID     string     `gorm:"column:trace_id;type:varchar(64)" json:"trace_id"` // 全流程追踪号
}

func (Hawthorn_task_execution) TableName() string {
	return "hawthorn_task_execution"
}
