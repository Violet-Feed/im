package message

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"im/dal"
	"im/proto_gen/im"
)

func StoreMessage(ctx context.Context, req *im.SaveMessageRequest) (resp *im.SaveMessageResponse, err error) {
	resp = &im.SaveMessageResponse{}
	convShortId := req.GetMsgBody().GetConvShortId()
	messageId := req.GetMsgBody().GetMsgId()
	key := fmt.Sprintf("msg:%d:%d", convShortId, messageId)
	messageBody, err := json.Marshal(req.GetMsgBody())
	if err != nil {
		logrus.Errorf("[StoreMessage] marshal messageBody err. err = %v", err)
		return nil, err
	}
	err = dal.KvrocksServer.Set(ctx, key, string(messageBody))
	if err != nil {
		logrus.Errorf("[StoreMessage] kvrocks set err. err = %v", err)
		return nil, err
	}
	return resp, nil
}
