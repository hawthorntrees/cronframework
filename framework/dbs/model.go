package dbs

import (
	"gorm.io/gorm"
)

var (
	dbManager = &DBManager{
		DBNameMap:  make(map[string]*DBInstance),
		DBGroupMap: make(map[string][]*DBInstance),
	}
)

type DBManager struct {
	DBNameMap  map[string]*DBInstance
	DBGroupMap map[string][]*DBInstance
}
type DBInstance struct {
	PrimaryDB      dbConnectInfo
	StandbyDB      dbConnectInfo
	CurrentDB      *dbConnectInfo
	Group          string
	IsUsingStandby bool
}
type dbConnectInfo struct {
	DB     *gorm.DB
	Center string
	Id     string
}

func getDBManager() *DBManager {
	return dbManager
}

type DBNameEnum string
type DBGroupEnum string

func GetDBInfoByName(name DBNameEnum) (db *gorm.DB, center string, group string, id string) {
	dbConnect := dbManager.DBNameMap[string(name)].CurrentDB
	if dbConnect.DB != nil {
		db = dbConnect.DB.Session(&gorm.Session{})
	} else {
		db = nil
	}
	return db, dbConnect.Center, dbManager.DBNameMap[string(name)].Group, dbConnect.Id
}

func GetDBByName(name DBNameEnum) (db *gorm.DB) {
	is, ok := dbManager.DBNameMap[string(name)]
	if ok {
		if is.CurrentDB != nil {
			return is.CurrentDB.DB.Session(&gorm.Session{})
		}
	}

	return nil
}

type dBInfo struct {
	Center string
	ID     string
	DB     *gorm.DB
	Group  string
}

func GetDBByGroup(group DBGroupEnum) (dbInfo []dBInfo) {
	instances, ok := dbManager.DBGroupMap[string(group)]
	if !ok {
		return nil
	}
	ret := make([]dBInfo, len(instances))

	for i, df := range instances {
		ret[i] = dBInfo{
			Center: df.CurrentDB.Center,
			ID:     df.CurrentDB.Id,
			DB:     df.CurrentDB.DB.Session(&gorm.Session{}),
			Group:  df.Group,
		}
	}
	return ret
}

func GetDB() (db *gorm.DB) {
	db = GetDBByName(DBNameEnum(serverDatabaseName))
	return db
}
func CloseDBS() {
	for _, instance := range dbManager.DBNameMap {
		db := instance.StandbyDB.DB
		if db == nil {
			break
		}
		sdb, err := db.DB()
		if err != nil {
			sdb.Close()
		}
	}
}
