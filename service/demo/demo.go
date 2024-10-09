package demo

import (
	"context"
	"fmt"
	demo "im/proto_gen"
)

type DemoService interface {
	GetMessage(ctx context.Context) (string, error)
}

type DemoServiceImpl struct {
	client demo.DemoClient
}

func NewDemoServiceImpl(client demo.DemoClient) DemoServiceImpl {
	return DemoServiceImpl{client: client}
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
