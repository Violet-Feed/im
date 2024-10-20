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

// TODO:get、ext、本地缓存、定期任务、监控
func main() {
	//ctx, cancel := context.WithCancel(context.Background())
	//sigCh := make(chan os.Signal, 1)
	//signal.Notify(sigCh, os.Interrupt)
	//go func() {
	//	defer cancel()
	//	<-sigCh
	//	logrus.Error("Received interrupt signal...")
	//}()
	//for {
	//	select {
	//	case <-ctx.Done():
	//		logrus.Error("Program interrupted...")
	//		return
	//	default:
	//		logrus.Info("Program is running...")
	//		time.Sleep(1 * time.Second)
	//	}
	//}
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
