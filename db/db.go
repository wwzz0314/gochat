package db

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var dbIns *gorm.DB

func init() {
	initDB("gochat")
}

func initDB(dbName string) {
	var e error
	dsn := fmt.Sprintf("root:123456@tcp(127.0.0.1:3306)/%s?charset=utf8&parseTime=True&loc=Local", dbName)
	db, e := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if e != nil {
		logrus.Errorf("mysql connect error: %s", e)
	}
	dbIns = db
	d, _ := db.DB()
	err := d.Ping()
	if e != nil {
		logrus.Errorf("mysql connect error: %s", err)
	}
}

func GetDb() (db *gorm.DB) {
	return dbIns
}

type DbGoChat struct {
}

func (d *DbGoChat) GetDbName() string {
	return "gochat"
}
