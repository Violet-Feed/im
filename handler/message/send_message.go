package message

import (
	"context"
	"github.com/sirupsen/logrus"
	"im/dal/mq"
	"im/handler/conversation"
	"im/proto_gen/im"
	"im/util"
	"strconv"
	"strings"
	"time"
)

func SendMessage(ctx context.Context, req *im.SendMessageRequest) (resp *im.SendMessageResponse, err error) {
	resp = &im.SendMessageResponse{
		BaseResp: &im.BaseResp{StatusCode: im.StatusCode_Success},
	}
	messageId := util.MsgIdGenerator.Generate().Int64()
	if req.GetConType() == int32(im.ConversationType_One_Chat) && req.GetConShortId() == 0 { //创建会话
		parts := strings.Split(req.GetConId(), ":")
		minId, _ := strconv.ParseInt(parts[0], 10, 64)
		maxId, _ := strconv.ParseInt(parts[1], 10, 64)
		createConversationRequest := &im.CreateConversationRequest{
			ConId:   req.ConId,
			ConType: req.ConType,
			Members: []int64{minId, maxId},
		}
		createConversationResponse, err := conversation.CreateConversation(ctx, createConversationRequest)
		if err != nil {
			logrus.Errorf("[SendMessage] CreateConversation err. err = %v", err)
			resp.BaseResp.StatusCode = im.StatusCode_Server_Error
			return resp, err
		}
		req.ConShortId = createConversationResponse.ConInfo.ConShortId
	}
	//TODO：消息频率控制
	createTime := time.Now().Unix()
	messageBody := &im.MessageBody{
		UserId:     req.UserId,
		ConId:      req.ConId,
		ConShortId: req.ConShortId,
		ConType:    req.ConType,
		MsgId:      util.Int64(messageId),
		MsgType:    req.MsgType,
		MsgContent: req.MsgContent,
		CreateTime: util.Int64(createTime),
	}
	messageEvent := &im.MessageEvent{
		MsgBody: messageBody,
	}
	err = mq.SendToMq(ctx, "conversation", strconv.FormatInt(req.GetConShortId(), 10), messageEvent)
	if err != nil {
		logrus.Errorf("[SendMessage] SendToMq err. err = %v", err)
		resp.BaseResp.StatusCode = im.StatusCode_Server_Error
		return resp, err
	}
	return resp, nil
}
