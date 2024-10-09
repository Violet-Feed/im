package dal

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	demo "im/proto_gen"
	"log"
)

type Dal struct {
	DemoClient    demo.DemoClient
	KvrocksClient *redis.Client
	MysqlClient   *gorm.DB
}

func InitClient() Dal {
	var dal Dal
	DemoCli, err := grpc.NewClient("127.0.0.1:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("无法连接到服务器：%v", err)
	} else {
		dal.DemoClient = demo.NewDemoClient(DemoCli)
	}

	dal.KvrocksClient = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6666",
		Password: "",
		DB:       0,
	})

	db, err := gorm.Open("mysql", "root:123456abc@tcp(127.0.0.1:3306)/im?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		panic("连接mysql失败, err=" + err.Error())
	} else {
		dal.MysqlClient = db
	}

	return dal
}
