package main

import (
	"context"
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	demo "im/proto_gen"
	"im/service"
	"log"
	"net"
)

func InitMq() {
	c, _ := rocketmq.NewPushConsumer(
		consumer.WithNameServer([]string{"127.0.0.1:9876"}),
		consumer.WithGroupName("test"),
	)
	if err := c.Subscribe("test", consumer.MessageSelector{}, func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for i := range msgs {
			fmt.Printf(string(msgs[i].Body))
		}
		return consumer.ConsumeSuccess, nil
	}); err != nil {
		fmt.Println(err.Error())
	}
	rlog.SetLogLevel("warn")
	err := c.Start()
	if err != nil {
		log.Fatalf("failed to mq consumer: %v", err)
	}
}

func main() {

	service.InitService()
	go func() {
		InitMq()
	}()

	go func() {
		r := gin.Default()
		r = Router(r)
		if err := r.Run(":9090"); err != nil {
			log.Fatalf("failed to http: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", ":9091")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	demo.RegisterDemoServer(s, &DemoServerImpl{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to rpc: %v", err)
	}
}
