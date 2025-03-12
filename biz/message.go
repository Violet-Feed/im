package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"im/dal"
	"im/dal/mq"
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
	//是否群成员
	var isMember int32
	if req.GetConType() == int32(im.ConversationType_One_Chat) {
		isMember = IsSingleMember(ctx, req.GetConId(), req.GetUserId())
	} else if req.GetConType() == int32(im.ConversationType_Group_Chat) {
		status, err := IsGroupsMember(ctx, []int64{req.GetConShortId()}, req.GetUserId())
		if err != nil {
			logrus.Errorf("[SendMessage] IsConversationMembers err. err = %v", err)
			resp.BaseResp.StatusCode = im.StatusCode_Server_Error
			return resp, err
		}
		isMember = status[0]
	}
	if isMember != 1 {
		resp.BaseResp.StatusCode = im.StatusCode_Not_Found_Error
		return resp, nil
	}
	if req.GetConType() == int32(im.ConversationType_One_Chat) && req.GetConShortId() == 0 { //创建会话
		parts := strings.Split(req.GetConId(), ":")
		minId, _ := strconv.ParseInt(parts[0], 10, 64)
		maxId, _ := strconv.ParseInt(parts[1], 10, 64)
		createConversationRequest := &im.CreateConversationRequest{
			ConId:   req.ConId,
			ConType: req.ConType,
			Members: []int64{minId, maxId},
		}
		createConversationResponse, err := CreateConversation(ctx, createConversationRequest)
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
		UserId:      req.UserId,
		ConId:       req.ConId,
		ConShortId:  req.ConShortId,
		ConType:     req.ConType,
		ClientMsgId: req.ClientMsgId,
		MsgId:       util.Int64(messageId),
		MsgType:     req.MsgType,
		MsgContent:  req.MsgContent,
		CreateTime:  util.Int64(createTime),
		Extra:       util.String(""),
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

func GetMessages(ctx context.Context, conShortId int64, msgIds []int64) ([]*im.MessageBody, error) {
	keys := make([]string, 0)
	for _, id := range msgIds {
		keys = append(keys, fmt.Sprintf("msg:%d:%d", conShortId, id))
	}
	messages, err := dal.KvrocksServer.MGet(ctx, keys)
	if err != nil {
		logrus.Errorf("[GetMessages] kvrocks mget err. err = %v", err)
		return nil, err
	}
	var messageBodies []*im.MessageBody
	for _, message := range messages {
		var messageBody im.MessageBody
		_ = json.Unmarshal([]byte(message), &messageBody)
		messageBodies = append(messageBodies, &messageBody)
	}
	return messageBodies, nil
}

func StoreMessage(ctx context.Context, msgBody *im.MessageBody) error {
	conShortId := msgBody.GetConShortId()
	messageId := msgBody.GetMsgId()
	key := fmt.Sprintf("msg:%d:%d", conShortId, messageId)
	messageBody, err := json.Marshal(msgBody)
	if err != nil {
		logrus.Errorf("[StoreMessage] marshal messageBody err. err = %v", err)
		return err
	}
	err = dal.KvrocksServer.Set(ctx, key, string(messageBody))
	if err != nil {
		logrus.Errorf("[StoreMessage] kvrocks set err. err = %v", err)
		return err
	}
	return nil
}
