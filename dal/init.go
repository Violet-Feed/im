package dal

import (
	"im/dal/kvrocks"
	"im/dal/mysql"
	"im/dal/redis"
	"im/dal/rpc"
)

var (
	DemoServer    rpc.DemoServiceImpl
	RedisServer   redis.RedisServiceImpl
	KvrocksServer kvrocks.KvrocksServiceImpl
	MysqlServer   mysql.MysqlServiceImpl
)

func InitService() {
	DemoServer = rpc.NewDemoServiceImpl()
	RedisServer = redis.NewRedisServiceImpl()
	KvrocksServer = kvrocks.NewKvrocksServiceImpl()
	MysqlServer = mysql.NewMysqlServiceImpl()
}
