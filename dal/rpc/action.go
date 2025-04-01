package rpc

import (
	"context"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"im/proto_gen/action"
)

type ActionService interface {
	MIsFollowing(ctx context.Context, fromUserId int64, toUserIds []int64) (map[int64]bool, error)
}

type ActionServiceImpl struct {
	client action.ActionServiceClient
}

func NewActionServiceImpl() ActionServiceImpl {
	actionServiceClient, err := grpc.NewClient("127.0.0.1:3003", grpc.WithTransportCredentials(insecure.NewCredentials()))
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
