package dal

import (
	"im/dal/demo"
	"im/dal/kvrocks"
	"im/dal/mysql"
	"im/dal/redis"
)

var (
	DemoServer    demo.DemoServiceImpl
	RedisServer   redis.RedisServiceImpl
	KvrocksServer kvrocks.KvrocksServiceImpl
	MysqlServer   mysql.MysqlServiceImpl
)

func InitService() {
	DemoServer = demo.NewDemoServiceImpl()
	RedisServer = redis.NewRedisServiceImpl()
	KvrocksServer = kvrocks.NewKvrocksServiceImpl()
	MysqlServer = mysql.NewMysqlServiceImpl()
}
