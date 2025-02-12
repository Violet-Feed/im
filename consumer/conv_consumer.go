package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/sirupsen/logrus"
	"im/dal/mq"
	"im/handler/conversation"
	"im/handler/index"
	"im/handler/message"
	"im/proto_gen/im"
	"im/util"
	"im/util/backoff"
	"strconv"
	"strings"
)

func ConvProcess(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	var messageEvent *im.MessageEvent
	_ = json.Unmarshal(msgs[0].Body, &messageEvent)
	logrus.Infof("[ConvProcess] rocketmq receive message success. message = %v", messageEvent)
	if !messageEvent.GetStored() && messageEvent.GetMsgBody().GetMsgType() < 1000 {
		saveMessageRequest := &im.SaveMessageRequest{
			MsgBody: messageEvent.GetMsgBody(),
		}
		_, err := message.StoreMessage(ctx, saveMessageRequest)
		if err != nil {
			logrus.Errorf("[ConvProcess] StoreMessage err. err = %v", err)
			retryCount := messageEvent.GetRetryCount()
			if retryCount > 3 {
				logrus.Errorf("[ConvProcess] retry too much. message = %v", messageEvent)
			}
			messageEvent.RetryCount = util.Int32(retryCount + 1)
			err = backoff.Retry(func() error {
				return mq.SendToMq(ctx, "conversation", strconv.FormatInt(messageEvent.GetMsgBody().GetConShortId(), 10), messageEvent)
			}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), backoff.Infinite))
			return consumer.ConsumeSuccess, nil
		}
		messageEvent.Stored = util.Bool(true)
		logrus.Infof("[ConvProcess] StoreMessage sucess")
	}
	if messageEvent.GetConIndex() == 0 && messageEvent.GetMsgBody().GetMsgType() < 1000 {
		appendConversationIndexRequest := &im.AppendConversationIndexRequest{
			ConShortId: messageEvent.GetMsgBody().ConShortId,
			MsgId:      messageEvent.GetMsgBody().MsgId,
		}
		appendConversationIndexResponse, err := index.AppendConversationIndex(ctx, appendConversationIndexRequest)
		if err != nil {
			logrus.Errorf("[ConvProcess] AppendConversationIndex err. err = %v", err)
			retryCount := messageEvent.GetRetryCount()
			if retryCount > 3 {
				logrus.Errorf("[ConvProcess] retry too much. message = %v", messageEvent)
				return consumer.ConsumeSuccess, nil
			}
			messageEvent.RetryCount = util.Int32(retryCount + 1)
			err = backoff.Retry(func() error {
				return mq.SendToMq(ctx, "conversation", strconv.FormatInt(messageEvent.GetMsgBody().GetConShortId(), 10), messageEvent)
			}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), backoff.Infinite))
			return consumer.ConsumeSuccess, nil
		}
		messageEvent.ConIndex = appendConversationIndexResponse.ConIndex
		logrus.Infof("[ConvProcess] AppendConversationIndex sucess. index = %v", messageEvent.GetConIndex())
	}
	receivers, err := getReceivers(ctx, messageEvent)
	if err != nil {
		retryCount := messageEvent.GetRetryCount()
		if retryCount > 3 {
			logrus.Errorf("[ConvProcess] retry too much. message = %v", messageEvent)
			return consumer.ConsumeSuccess, nil
		}
		messageEvent.RetryCount = util.Int32(retryCount + 1)
		err = backoff.Retry(func() error {
			return mq.SendToMq(ctx, "conversation", strconv.FormatInt(messageEvent.GetMsgBody().GetConShortId(), 10), messageEvent)
		}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), backoff.Infinite))
		return consumer.ConsumeSuccess, nil
	}
	for _, receiver := range receivers {
		err := mq.SendToMq(ctx, "user", strconv.FormatInt(receiver, 10), messageEvent)
		if err != nil {
		}
	}
	return consumer.ConsumeSuccess, nil
	//消费消息(message_parallel_consumer)->校验过滤->重试消息检查消息是否存在(message_api)->获取strategies->callback->处理ext信息
	//->写会话链(inbox_api,V2)->保存消息->(增加thread未读)->写入用户消息队列->与同步MQ互补->失败发送backup队列
}

func getReceivers(ctx context.Context, event *im.MessageEvent) ([]int64, error) {
	switch event.GetMsgBody().GetConType() {
	case int32(im.ConversationType_One_Chat):
		parts := strings.Split(event.GetMsgBody().GetConId(), ":")
		minId, _ := strconv.ParseInt(parts[0], 10, 64)
		maxId, _ := strconv.ParseInt(parts[1], 10, 64)
		return []int64{minId, maxId}, nil
	case int32(im.ConversationType_Group_Chat):
		getConversationMembersRequest := &im.GetConversationMembersRequest{
			ConShortId: event.GetMsgBody().ConShortId,
			OnlyId:     util.Bool(true),
		}
		getConversationMembersResponse, err := conversation.GetConversationMembers(ctx, getConversationMembersRequest)
		if err != nil {
			logrus.Errorf("[ConvProcess] GetConversationMembers err. err = %v", err)
			return nil, err
		}
		var receivers []int64
		for _, info := range getConversationMembersResponse.UserInfos {
			receivers = append(receivers, info.GetUserId())
		}
		return receivers, nil
	default:
		return nil, errors.New("conversation type invalid")
	}
}
