package mysql

import (
	"im/conf"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/sirupsen/logrus"
)

func NewMysqlDB(cfg conf.MySQLConfig) *gorm.DB {
	db, err := gorm.Open("mysql", cfg.DSN())
	if err != nil {
		logrus.Fatalf("[NewMysqlDB] mysql connect err. err = %v", err)
	}
	db.LogMode(false)
	return db
}
