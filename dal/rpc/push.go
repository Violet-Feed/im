package rpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"im/proto_gen/common"
	"im/proto_gen/push"
)

type PushService interface {
	Push(ctx context.Context, req *push.PushRequest) error
}

type PushServiceImpl struct {
	client push.PushServiceClient
}

func NewPushServiceImpl() PushServiceImpl {
	pushServiceClient, err := grpc.NewClient("127.0.0.1:3002", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logrus.Fatalf("[NewPushServiceImpl] rpc connect err. err = %v", err)
	}
	return PushServiceImpl{client: push.NewPushServiceClient(pushServiceClient)}
}

func (p *PushServiceImpl) Push(ctx context.Context, req *push.PushRequest) error {
	resp, err := p.client.Push(ctx, req)
	if err != nil {
		logrus.Errorf("[Push] rpc err. err = %v", err)
		return err
	}
	if resp.GetBaseResp().GetStatusCode() != common.StatusCode_Success {
		logrus.Errorf("[Push] rpc status err. err = %v", resp.GetBaseResp().GetStatusCode())
		return errors.New(fmt.Sprintf("rpc status err: %v", resp.GetBaseResp().GetStatusCode()))
	}
	return nil
}
