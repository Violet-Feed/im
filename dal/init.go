package dal

import (
	"github.com/jinzhu/gorm"
	"im/dal/kvrocks"
	"im/dal/mysql"
	"im/dal/redis"
	"im/dal/rpc"
)

var (
	DemoServer    rpc.DemoServiceImpl
	RedisServer   redis.RedisServiceImpl
	KvrocksServer kvrocks.KvrocksServiceImpl
	MysqlDB       *gorm.DB
)

func InitService() {
	DemoServer = rpc.NewDemoServiceImpl()
	RedisServer = redis.NewRedisServiceImpl()
	KvrocksServer = kvrocks.NewKvrocksServiceImpl()
	MysqlDB = mysql.NewMysqlDB()
}
