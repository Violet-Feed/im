package main

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"im/consumer"
	"im/dal"
	"im/proto_gen/im"
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
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
		ForceColors:     true,
	})
	dal.InitService()
	consumer.InitConsumer()
	go func() {
		r := gin.Default()
		r = Router(r)
		if err := r.Run(":3005"); err != nil {
			logrus.Fatalf("[main] gin run err. err = %v", err)
		}
	}()

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
