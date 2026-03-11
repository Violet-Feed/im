package dal

import (
	"im/dal/kvrocks"
	"im/dal/mysql"
	"im/dal/redis"
	"im/dal/rpc"

	"github.com/jinzhu/gorm"
)

var (
	PushServer    rpc.PushServiceImpl
	ActionServer  rpc.ActionServiceImpl
	RedisServer   redis.RedisServiceImpl
	KvrocksServer kvrocks.KvrocksServiceImpl
	MysqlDB       *gorm.DB
)

func InitService() {
	PushServer = rpc.NewPushServiceImpl()
	ActionServer = rpc.NewActionServiceImpl()
	RedisServer = redis.NewRedisServiceImpl()
	KvrocksServer = kvrocks.NewKvrocksServiceImpl()
	MysqlDB = mysql.NewMysqlDB()
}
