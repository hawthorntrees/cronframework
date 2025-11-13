package config

import (
	"reflect"
	"time"
)

func completeServer(c *Config) {
	if c.Server.Address == "" {
		c.Server.Address = ":8080"
	}
	if c.Server.BashPath == "" {
		c.Server.BashPath = "/"
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 10 * time.Second
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 10 * time.Second
	}
	if c.Server.IdleTimeout == 0 {
		c.Server.IdleTimeout = 30 * time.Second
	}
	if c.Server.SessionExpires == 0 {
		c.Server.SessionExpires = 24 * time.Hour
	}
}
func completeLog(c *Config) {
	if c.Logger.Filename == "" {
		c.Logger.Filename = "logs/app.log"
	}
	if c.Logger.MaxSize == 0 {
		c.Logger.MaxSize = 10 // 10MB
	}
	if c.Logger.MaxFiles == 0 {
		c.Logger.MaxFiles = 10 // 最多10个文件轮转
	}
	if c.Logger.MaxAge == 0 {
		c.Logger.MaxAge = 7
	}
}

func completeCronTask(c *Config) {
	if c.CronTask.TaskSyncInterval == 0 {
		c.CronTask.TaskSyncInterval = 15 * time.Second
	}
	if c.CronTask.NodeID == "" {
		c.CronTask.NodeID = c.App.Name
	}
}

func completeDatabases(c *Config) {
	for _, dbcfg := range c.Databases.ListsConfig {
		mergeConfig(&dbcfg.PrimaryConfig, &dbcfg.DefaultConfig, &c.Databases.DefaultConfig)
		mergeConfig(&dbcfg.StandbyConfig, &dbcfg.DefaultConfig, &c.Databases.DefaultConfig)
	}
}
func mergeConfig(dst interface{}, srcs ...interface{}) {
	if dst == nil {
		return
	}

	dstVal := reflect.ValueOf(dst).Elem()
	numFields := dstVal.NumField()

	var srcVals []reflect.Value
	for _, src := range srcs {
		if src != nil {
			srcVals = append(srcVals, reflect.ValueOf(src).Elem())
		}
	}

	for i := 0; i < numFields; i++ {
		dstField := dstVal.Field(i)
		if !dstField.IsZero() {
			continue
		}

		for _, srcVal := range srcVals {
			srcField := srcVal.Field(i)
			if !srcField.IsZero() {
				dstField.Set(srcField)
				break
			}
		}
	}
}
