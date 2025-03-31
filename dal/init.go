package dal

import (
	"github.com/jinzhu/gorm"
	"im/dal/kvrocks"
	"im/dal/mysql"
	"im/dal/redis"
	"im/dal/rpc"
)

var (
	ActionServer  rpc.ActionServiceImpl
	RedisServer   redis.RedisServiceImpl
	KvrocksServer kvrocks.KvrocksServiceImpl
	MysqlDB       *gorm.DB
)

func InitService() {
	ActionServer = rpc.NewActionServiceImpl()
	RedisServer = redis.NewRedisServiceImpl()
	KvrocksServer = kvrocks.NewKvrocksServiceImpl()
	MysqlDB = mysql.NewMysqlDB()
}
