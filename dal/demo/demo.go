package demo

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	demo "im/proto_gen"
	"log"
)

type DemoService interface {
	GetMessage(ctx context.Context) (string, error)
}

type DemoServiceImpl struct {
	client demo.DemoClient
}

func NewDemoServiceImpl() DemoServiceImpl {
	demoClient, err := grpc.NewClient("127.0.0.1:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("[NewDemoServiceImpl] rpc connect err. err = %v", err)
	}
	return DemoServiceImpl{client: demo.NewDemoClient(demoClient)}
}

func (d *DemoServiceImpl) GetMessage(ctx context.Context, param string) (string, error) {
	req := &demo.DemoRequest{
		Param: param,
	}
	resp, err := d.client.GetMessage(ctx, req)
	if err != nil {
		fmt.Println(resp, err)
		return "", err
	}
	return resp.GetMessage(), nil
}
