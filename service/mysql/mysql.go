package mysql

import (
	"github.com/jinzhu/gorm"
)

type MysqlService interface {
}

type MysqlServiceImpl struct {
	client *gorm.DB
}

func NewMysqlServiceImpl(client *gorm.DB) MysqlServiceImpl {
	return MysqlServiceImpl{client: client}
}
