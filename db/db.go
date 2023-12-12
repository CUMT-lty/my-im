package db

import (
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"sync"
	"time"
)

// 数据库连接管理
var dbMap = map[string]*gorm.DB{}

// 锁
var syncLock sync.Mutex

func init() {
	initDB("gochat")
}

func initDB(dbName string) {
	var e error
	// if prod env , you should change mysql driver for yourself !!!
	dsn := "root:@tcp(127.0.0.1:3306)/gochat?charset=utf8mb4&parseTime=True&loc=Local"
	var newLogger = logger.New( // 自定义日志模板，打印 sql 语句
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		})
	syncLock.Lock() // TODO: 加锁，锁定对数据库连接的操作，并发安全
	dbMap[dbName], e = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	// TODO: 这些在 mysql 中都对应什么
	//dbMap[dbName].DB().SetMaxIdleConns(4)
	//dbMap[dbName].DB().SetMaxOpenConns(20)
	//dbMap[dbName].DB().SetConnMaxLifetime(8 * time.Second)
	//if config.GetMode() == "dev" {
	//	dbMap[dbName].LogMode(true)
	//}
	syncLock.Unlock() // TODO: 释放锁
	if e != nil {
		logrus.Error("connect db fail:%s", e.Error())
	}
}

func GetDb(dbName string) (db *gorm.DB) {
	return dbMap[dbName]
}
