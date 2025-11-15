package config

import "time"

type Config struct {
	App       AppConfig       `yaml:"app,omitempty"`
	Server    ServerConfig    `yaml:"server,omitempty"`
	Logger    LoggerConfig    `yaml:"logger,omitempty"`
	CronTask  TaskConfig      `yaml:"cronTask,omitempty"`
	Databases DatabasesConfig `yaml:"databases,omitempty"`
}

type AppConfig struct {
	Name   string `yaml:"name,omitempty"`
	Center string `yaml:"center,omitempty"`
	Group  string `yaml:"group,omitempty"`
	Id     string `yaml:"id,omitempty"`
	WorkID uint16 `yaml:"workID,omitempty"`
}

type ServerConfig struct {
	BashPath       string        `yaml:"bash_path,omitempty"`
	Address        string        `yaml:"address,omitempty"`
	ReadTimeout    time.Duration `yaml:"read_timeout,omitempty"`
	WriteTimeout   time.Duration `yaml:"write_timeout,omitempty"`
	IdleTimeout    time.Duration `yaml:"idle_timeout,omitempty"`
	DatabaseName   string        `yaml:"database_name,omitempty"`
	TokenKey       string        `yaml:"token_key,omitempty"`
	SessionExpires time.Duration `yaml:"session_expires"`
}

type TaskConfig struct {
	NodeID                 string        `yaml:"node_id,omitempty"`
	LogLevel               string        `yaml:"log_level,omitempty"`
	TaskSyncInterval       time.Duration `yaml:"task_sync_interval,omitempty"`
	NotRecordTaskExecution bool          `yaml:"not_record_task_execution"`
}

type LoggerConfig struct {
	Level    string `yaml:"level,omitempty"`
	OrmLevel string `yaml:"ormLevel,omitempty"`
	Filename string `yaml:"filename,omitempty"`
	MaxSize  int    `yaml:"maxsize,omitempty"`
	MaxFiles int    `yaml:"maxFiles,omitempty"`
	MaxAge   int    `yaml:"maxAge,omitempty"` // 日志最大保留的天数
	Compress bool   `yaml:"compress,omitempty"`
}

type DatabasesConfig struct {
	DatabaseHealthMonitorInterval time.Duration              `yaml:"database_health_monitor_interval,omitempty"`
	DefaultConfig                 DBInstanceConfig           `yaml:"default,omitempty"`
	ListsConfig                   map[string]*DatabaseConfig `yaml:"lists,omitempty"`
}

type DatabaseConfig struct {
	Cluster       string           `yaml:"cluster,omitempty"`
	DefaultConfig DBInstanceConfig `yaml:"default,omitempty"`
	PrimaryConfig DBInstanceConfig `yaml:"primary,omitempty"`
	StandbyConfig DBInstanceConfig `yaml:"standby,omitempty"`
}

type DBInstanceConfig struct {
	DSN             string        `yaml:"dsn,omitempty"`
	Center          string        `yaml:"center,omitempty"`
	Id              string        `yaml:"id,omitempty"`
	Schema          string        `yaml:"schema,omitempty"`
	Password        string        `yaml:"password,omitempty"`
	MaxOpenConns    int           `yaml:"max_open_conns,omitempty"`
	MaxIdleConns    int           `yaml:"max_idle_conns,omitempty"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime,omitempty"`
	SM4Key          string        `yaml:"sm4_key,omitempty"`
}
