package dbs

import (
	"gorm.io/gorm"
)

var (
	dbManager = &DBManager{
		DBNameMap:    make(map[string]*DBInstance),
		DBClusterMap: make(map[string][]*DBInstance),
	}
)

type DBManager struct {
	DBNameMap    map[string]*DBInstance
	DBClusterMap map[string][]*DBInstance
}
type DBInstance struct {
	PrimaryDB      dbConnectInfo
	StandbyDB      dbConnectInfo
	CurrentDB      *dbConnectInfo
	Cluster        string
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
type DBClusterEnum string

func GetDBInfoByName(name DBNameEnum) (db *gorm.DB, center string, cluster string, id string) {
	dbConnect := dbManager.DBNameMap[string(name)].CurrentDB
	if dbConnect.DB != nil {
		db = dbConnect.DB.Session(&gorm.Session{})
	} else {
		db = nil
	}
	return db, dbConnect.Center, dbManager.DBNameMap[string(name)].Cluster, dbConnect.Id
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
	Center  string
	ID      string
	DB      *gorm.DB
	Cluster string
}

func GetDBByCluster(cluster DBClusterEnum) (dbInfo []dBInfo) {
	instances, ok := dbManager.DBClusterMap[string(cluster)]
	if !ok {
		return nil
	}
	ret := make([]dBInfo, len(instances))

	for i, df := range instances {
		ret[i] = dBInfo{
			Center:  df.CurrentDB.Center,
			ID:      df.CurrentDB.Id,
			DB:      df.CurrentDB.DB.Session(&gorm.Session{}),
			Cluster: df.Cluster,
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
