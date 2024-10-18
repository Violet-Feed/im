package push

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"im/dal"
	"im/proto_gen/im"
)

func Push(ctx context.Context, req *im.PushRequest) (resp *im.PushResponse, err error) {
	resp = &im.PushResponse{}
	//TODO:通过userId在redis HGet所有ConnectionId,通过ConnectionId得到连接（how？）
	userId := req.GetMsgBody().GetUserId()
	key := fmt.Sprintf("onlineUser:%d", userId)
	connections, err := dal.RedisServer.HGetAll(ctx, key)
	if err != nil {
		logrus.Errorf("[Push] redid HGetAll err. err = %v", err)
		return nil, err
	}
	for id, info := range connections {
		logrus.Infof("[Push] get connections. id = %v, info = %v", id, info)
	}
	return resp, nil
}
