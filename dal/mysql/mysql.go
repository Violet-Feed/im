package mysql

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/sirupsen/logrus"
)

var DB *gorm.DB

func InitMysql() {
	db, err := gorm.Open("mysql", "root:123456abc@tcp(127.0.0.1:3306)/im?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		logrus.Fatalf("[NewMysqlServiceImpl] mysql connect err. err = %v", err)
	}
	db.LogMode(false)
	DB = db
}
