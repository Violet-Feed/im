package main

import (
	"context"
	demo "im/proto_gen"
)

type DemoServerImpl struct {
	demo.UnimplementedDemoServer
}

func (s *DemoServerImpl) GetMessage(ctx context.Context, req *demo.DemoRequest) (*demo.DemoResponse, error) {
	// 实现具体的逻辑
	param := req.GetParam()
	if param == "1" {
		return &demo.DemoResponse{Message: "Hello from implemented server"}, nil
	} else {
		return &demo.DemoResponse{Message: "GoodBye from implemented server"}, nil
	}
}
