package rpc

import (
	"context"
	"im/conf"
	"im/proto_gen/action"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ActionService interface {
	MIsFollowing(ctx context.Context, fromUserId int64, toUserIds []int64) (map[int64]bool, error)
}

type ActionServiceImpl struct {
	client action.ActionServiceClient
}

func NewActionServiceImpl(cfg conf.RPCConfig) ActionServiceImpl {
	actionServiceClient, err := grpc.NewClient(cfg.ActionAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logrus.Fatalf("[NewActionServiceImpl] rpc connect err. err = %v", err)
	}
	return ActionServiceImpl{client: action.NewActionServiceClient(actionServiceClient)}
}

func (a *ActionServiceImpl) MIsFollowing(ctx context.Context, fromUserId int64, toUserIds []int64) (map[int64]bool, error) {
	req := &action.MIsFollowRequest{
		FromUserId: fromUserId,
		ToUserIds:  toUserIds,
	}
	resp, err := a.client.MIsFollowing(ctx, req)
	if err != nil {
		logrus.Errorf("[MIsFollowing] rpc err. err = %v", err)
		return nil, err
	}
	return resp.GetIsFollowing(), nil
}
