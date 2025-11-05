package dbs

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/hawthorntrees/cronframework/framework/config"
	"github.com/hawthorntrees/cronframework/framework/utils"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

var serverDatabaseName string

func Init(log *zap.Logger, cfg *config.Config) {
	initLogger(cfg.Logger.OrmLevel)
	serverDatabaseName = cfg.Server.DatabaseName
	manager := getDBManager()
	for name, dbcfg := range cfg.Databases.ListsConfig {
		instance, msg := dbInit(name, dbcfg)
		manager.DBNameMap[name] = instance
		manager.DBGroupMap[dbcfg.Group] = append(manager.DBGroupMap[dbcfg.Group], instance)
		log.Debug(msg)
	}
	if cfg.Databases.DatabaseHealthMonitorInterval > 0 {
		go startHealthCheck(&cfg.Databases)
	}
}

func dbInit(name string, cfg *config.DatabaseConfig) (*DBInstance, string) {
	logmsg := name
	instance := &DBInstance{
		PrimaryDB:      dbConnectInfo{},
		StandbyDB:      dbConnectInfo{},
		CurrentDB:      nil,
		Group:          "",
		IsUsingStandby: false,
	}
	priDB, err := initDBConnection(&cfg.PrimaryConfig)
	if err != nil {
		logmsg = logmsg + " 初始化主库失败;"
	} else {
		instance.PrimaryDB.Center = cfg.PrimaryConfig.Center
		instance.PrimaryDB.Id = cfg.PrimaryConfig.Id
		instance.PrimaryDB.DB = priDB
		instance.CurrentDB = &instance.PrimaryDB
		logmsg = logmsg + " 初始化主库成功;"
	}

	if cfg.StandbyConfig.DSN != "" {
		staDB, err := initDBConnection(&cfg.StandbyConfig)
		if err != nil {
			logmsg = logmsg + " 初始化备库失败;"
		} else {
			instance.StandbyDB.Center = cfg.StandbyConfig.Center
			instance.StandbyDB.Id = cfg.StandbyConfig.Id
			instance.StandbyDB.DB = staDB
			if instance.CurrentDB == nil {
				instance.CurrentDB = &instance.StandbyDB
				instance.IsUsingStandby = true
			}
			logmsg = logmsg + " 初始化备库成功;"
		}
	}
	if instance.CurrentDB == nil {
		panic("主备库都无法连接，启动失败:" + name)
	}
	return instance, logmsg
}

// 没有log，但是数据库初始化时，要传入gorm的logger，所以，gorm的日志初始化，放在logger里面还是放在dbs里面呢？纠结啊
func initDBConnection(instance *config.DBInstanceConfig) (*gorm.DB, error) {
	decryptedPwd, err := utils.SM4Decrypt([]byte(instance.SM4Key), instance.Password)
	if err != nil {
		return nil, fmt.Errorf("解密密码失败: %w", err)
	}
	dsn := fmt.Sprintf(instance.DSN, decryptedPwd)
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	poolCfg.MaxConns = int32(instance.MaxOpenConns)
	poolCfg.MinConns = int32(instance.MaxIdleConns)
	poolCfg.MaxConnLifetime = instance.ConnMaxLifetime
	poolCfg.MaxConnIdleTime = 5 * time.Minute
	//poolCfg.HealthCheckPeriod = 10 * time.Second
	poolCfg.ConnConfig.ConnectTimeout = 5 * time.Second // 连接超时时间
	poolCfg.ConnConfig.RuntimeParams["search_path"] = instance.Schema
	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}
	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, err
	}
	connector := stdlib.GetPoolConnector(pool)
	sqlDB := sql.OpenDB(connector)
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		PrepareStmt: true,
		Logger:      zaplog,
	})
	return db, nil
}
