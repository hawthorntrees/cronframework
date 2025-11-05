package dbs

import (
	"github.com/hawthorntrees/cronframework/framework/config"
	"gorm.io/gorm"
	"time"
)

func startHealthCheck(cfg *config.DatabasesConfig) {
	manager := getDBManager()
	listsConfig := cfg.ListsConfig

	ticker := time.NewTicker(cfg.DatabaseHealthMonitorInterval * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		for name, instance := range manager.DBNameMap {
			checkAndSwitch(instance, listsConfig[name])
		}
	}
}

func checkAndSwitch(dbIns *DBInstance, cfg *config.DatabaseConfig) {
	if isHealthy(dbIns.CurrentDB.DB) {
		return
	}
	if dbIns.IsUsingStandby {
		if dbIns.PrimaryDB.DB == nil {
			dbIns.PrimaryDB.DB, _ = initDBConnection(&cfg.PrimaryConfig)
		}
		if isHealthy(dbIns.PrimaryDB.DB) {
			dbIns.CurrentDB = &dbIns.PrimaryDB
			dbIns.IsUsingStandby = false
		}
	} else {
		if cfg.StandbyConfig.DSN == "" {
			return
		}
		if dbIns.StandbyDB.DB == nil {
			dbIns.StandbyDB.DB, _ = initDBConnection(&cfg.PrimaryConfig)
		}
		if isHealthy(dbIns.StandbyDB.DB) {
			dbIns.CurrentDB = &dbIns.StandbyDB
			dbIns.IsUsingStandby = true
		}
	}
}

func isHealthy(db *gorm.DB) bool {
	if db == nil {
		return false
	}
	sqlDB, err := db.DB()
	if err != nil {
		return false
	}
	return sqlDB.Ping() == nil
}
