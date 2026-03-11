package main

import (
	"im/consumer"
	"im/dal"
	"im/proto_gen/im"
	"net"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		ForceColors:     true,
	})
	logrus.Infof("[main] server start")
	dal.InitService()
	consumer.InitConsumer()
	lis, err := net.Listen("tcp", ":3004")
	if err != nil {
		logrus.Fatalf("[main] grpc listen err. err = %v", err)
	}
	s := grpc.NewServer()
	im.RegisterIMServiceServer(s, &IMServerImpl{})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		logrus.Fatalf("[main] grpc run err. err = %v", err)
	}
}
