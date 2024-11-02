package message

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"im/dal"
	"im/proto_gen/im"
)

func GetMessages(ctx context.Context, req *im.GetMessagesRequest) (resp *im.GetMessagesResponse, err error) {
	resp = &im.GetMessagesResponse{
		BaseResp: &im.BaseResp{StatusCode: im.StatusCode_Success},
	}
	conShortId := req.GetConShortId()
	messageIds := req.GetMsgIds()
	keys := make([]string, len(messageIds))
	for _, id := range messageIds {
		keys = append(keys, fmt.Sprintf("%d:%d", conShortId, id))
	}
	messages, err := dal.KvrocksServer.MGet(ctx, keys)
	if err != nil {
		logrus.Errorf("[GetMessages] kvrocks mget err. err = %v", err)
		resp.BaseResp.StatusCode = im.StatusCode_Server_Error
		return nil, err
	}
	var messageBodies []*im.MessageBody
	for _, message := range messages {
		var messageBody im.MessageBody
		err = json.Unmarshal([]byte(message), &messageBody)
		if err != nil {
			logrus.Errorf("[GetMessages] unmarshal messageBody err. msg = %v, err = %v", message, err)
			//TODO
		} else {
			messageBodies = append(messageBodies, &messageBody)
		}
	}
	resp.MsgBodies = messageBodies
	return resp, nil
}
