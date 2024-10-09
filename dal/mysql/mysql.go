package mysql

import (
	"github.com/jinzhu/gorm"
	"log"
)

type MysqlService interface {
}

type MysqlServiceImpl struct {
	client *gorm.DB
}

func NewMysqlServiceImpl() MysqlServiceImpl {
	db, err := gorm.Open("mysql", "root:123456abc@tcp(127.0.0.1:3306)/im?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatalf("[NewMysqlServiceImpl] mysql connect err. err = %v", err)
	}
	return MysqlServiceImpl{client: db}
}
