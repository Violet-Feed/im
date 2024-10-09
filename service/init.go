package service

import (
	"im/dal"
	"im/service/demo"
	"im/service/kvrocks"
	"im/service/mysql"
)

var (
	DemoServer    demo.DemoServiceImpl
	KvrocksServer kvrocks.KvrocksServiceImpl
	MysqlServer   mysql.MysqlServiceImpl
)

func InitService() {
	client := dal.InitClient()

	DemoServer = demo.NewDemoServiceImpl(client.DemoClient)
	KvrocksServer = kvrocks.NewKvrocksServiceImpl(client.KvrocksClient)
	MysqlServer = mysql.NewMysqlServiceImpl(client.MysqlClient)
}
