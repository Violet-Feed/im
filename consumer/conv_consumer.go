package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/sirupsen/logrus"
	"im/biz"
	"im/biz/constant"
	"im/dal/mq"
	"im/proto_gen/im"
	"strconv"
	"strings"
)

func ConvProcess(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	var messageEvent *im.MessageEvent
	_ = json.Unmarshal(msgs[0].Body, &messageEvent)
	logrus.Infof("[ConvProcess] rocketmq receive message success. message = %v", messageEvent)
	//保存消息
	if !messageEvent.GetStored() {
		err := biz.StoreMessage(ctx, messageEvent.GetMsgBody())
		if err != nil {
			logrus.Errorf("[ConvProcess] StoreMessage err. err = %v", err)
			return mq.SendToRetry(ctx, constant.IM_CONV_TOPIC, messageEvent)
		}
		messageEvent.Stored = true
	}
	//写入会话链
	if messageEvent.GetConIndex() == 0 && messageEvent.GetMsgBody().GetMsgType() < 100 {
		conIndex, err := biz.AppendConversationIndex(ctx, messageEvent.GetMsgBody().GetConShortId(), messageEvent.GetMsgBody().GetMsgId())
		if err != nil {
			logrus.Errorf("[ConvProcess] AppendConversationIndex err. err = %v", err)
			return mq.SendToRetry(ctx, constant.IM_CONV_TOPIC, messageEvent)
		}
		messageEvent.ConIndex = conIndex
		messageEvent.MsgBody.ConIndex = conIndex
	}
	//获取成员分发用户topic
	receivers, err := getReceivers(ctx, messageEvent)
	if err != nil {
		return mq.SendToRetry(ctx, constant.IM_CONV_TOPIC, messageEvent)
	}
	for _, receiver := range receivers {
		err := mq.SendToMq(ctx, constant.IM_USER_TOPIC, strconv.FormatInt(receiver, 10), messageEvent)
		if err != nil {
			logrus.Errorf("[ConvProcess] SendToMq err. err = %v", err)
		}
	}
	return consumer.ConsumeSuccess, nil
	//消费消息(message_parallel_consumer)->校验过滤->重试消息检查消息是否存在(message_api)->获取strategies->callback->处理ext信息
	//->写会话链(inbox_api,V2)->保存消息->(增加thread未读)->写入用户消息队列->与同步MQ互补->失败发送backup队列
}

func getReceivers(ctx context.Context, event *im.MessageEvent) ([]int64, error) {
	if event.GetMsgBody().GetMsgType() == int32(im.MessageType_MarkRead) {
		return []int64{event.GetMsgBody().GetUserId()}, nil
	}
	switch event.GetMsgBody().GetConType() {
	case int32(im.ConversationType_One_Chat):
		parts := strings.Split(event.GetMsgBody().GetConId(), ":")
		minId, _ := strconv.ParseInt(parts[0], 10, 64)
		maxId, _ := strconv.ParseInt(parts[1], 10, 64)
		return []int64{minId, maxId}, nil
	case int32(im.ConversationType_Group_Chat):
		receivers, err := biz.GetConversationMemberIds(ctx, event.GetMsgBody().GetConShortId())
		if err != nil {
			logrus.Errorf("[ConvProcess] GetConversationMembers err. err = %v", err)
			return nil, err
		}
		return receivers, nil
	case int32(im.ConversationType_AI_Chat):
		parts := strings.Split(event.GetMsgBody().GetConId(), ":")
		userId, _ := strconv.ParseInt(parts[1], 10, 64)
		return []int64{userId}, nil
	default:
		return nil, errors.New("conversation type invalid")
	}
}
