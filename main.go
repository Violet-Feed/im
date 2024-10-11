package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"im/consumer"
	"im/dal"
	demo "im/proto_gen"
	"net"
)

func main() {
	dal.InitService()
	consumer.InitConsumer()
	go func() {
		r := gin.Default()
		r = Router(r)
		if err := r.Run(":9090"); err != nil {
			logrus.Fatalf("[main] gin run err. err = %v", err)
		}
	}()

	lis, err := net.Listen("tcp", ":9091")
	if err != nil {
		logrus.Fatalf("[main] grpc listen err. err = %v", err)
	}
	s := grpc.NewServer()
	demo.RegisterDemoServer(s, &DemoServerImpl{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		logrus.Fatalf("[main] grpc run err. err = %v", err)
	}
}
