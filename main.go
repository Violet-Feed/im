package main

import (
	"context"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"im/dal"
	demo "im/proto_gen"
	"log"
	"net"
)

func initRocketMq() {
	c, _ := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{"127.0.0.1:9876"}),
		consumer.WithGroupName("test"),
	)
	if err := c.Subscribe("test", consumer.MessageSelector{}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for i := range msgs {
			log.Printf("[initRocketMq] rocketmq consume message success. message = %v", msgs[i].Body)
		}
		return consumer.ConsumeSuccess, nil
	}); err != nil {
		log.Printf("[initRocketMq] rocketmq consume message err. err = %v", err)
	}
	rlog.SetLogLevel("warn")
	err := c.Start()
	if err != nil {
		log.Fatalf("[initRocketMq] rocketmq consumer run err. err = %v", err)
	}
}

func main() {

	dal.InitService()

	go func() {
		initRocketMq()
	}()

	go func() {
		r := gin.Default()
		r = Router(r)
		if err := r.Run(":9090"); err != nil {
			log.Fatalf("[main] gin run err. err = %v", err)
		}
	}()

	lis, err := net.Listen("tcp", ":9091")
	if err != nil {
		log.Fatalf("[main] grpc listen err. err = %v", err)
	}
	s := grpc.NewServer()
	demo.RegisterDemoServer(s, &DemoServerImpl{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("[main] grpc run err. err = %v", err)
	}
}
