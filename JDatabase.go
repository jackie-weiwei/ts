package ts

import (
	"container/list"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type FuncCreateTable func()

var (
	DB     *gorm.DB
	tables *list.List = list.New()
)

func RegisterTable(createTableFunc FuncCreateTable) {
	tables.PushBack(createTableFunc)
}

func InitDatabase(dbUser, dbPasswd, dbHost, dbPort, dbName string) bool {
	dsn := dbUser + ":" + dbPasswd + "@(" + dbHost + ":" + dbPort + ")/" + dbName + "?charset=utf8mb4&parseTime=True&loc=UTC"
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent), NowFunc: func() time.Time { return time.Now().UTC() }})

	if err != nil {
		return false
	}
	createTables()
	return true
}

func createTables() {
	for e := tables.Front(); e != nil; e = e.Next() {
		c := e.Value.(FuncCreateTable)
		c()
	}
}
