package dal

import (
	"im/conf"
	"im/dal/kvrocks"
	"im/dal/mq"
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
	PushServer = rpc.NewPushServiceImpl(conf.C.RPC)
	ActionServer = rpc.NewActionServiceImpl(conf.C.RPC)
	RedisServer = redis.NewRedisServiceImpl(conf.C.Redis)
	KvrocksServer = kvrocks.NewKvrocksServiceImpl(conf.C.Kvrocks)
	MysqlDB = mysql.NewMysqlDB(conf.C.MySQL)
	mq.InitProducer(conf.C.RocketMQ)
}