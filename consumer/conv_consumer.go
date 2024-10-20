package consumer

import (
	"context"
	"encoding/json"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/sirupsen/logrus"
	"im/dal/mq"
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
				return mq.SendToMq(ctx, "conversation", strconv.FormatInt(messageEvent.GetMsgBody().GetConvShortId(), 10), messageEvent)
			}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), backoff.Infinite))
			return consumer.ConsumeSuccess, nil
		}
		messageEvent.Stored = util.Bool(true)
		logrus.Infof("[ConvProcess] StoreMessage sucess")
	}
	if messageEvent.GetConvIndex() == 0 && messageEvent.GetMsgBody().GetMsgType() < 1000 {
		appendConversationIndexRequest := &im.AppendConversationIndexRequest{
			ConvShortId: messageEvent.GetMsgBody().ConvShortId,
			MsgId:       messageEvent.GetMsgBody().MsgId,
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
				return mq.SendToMq(ctx, "conversation", strconv.FormatInt(messageEvent.GetMsgBody().GetConvShortId(), 10), messageEvent)
			}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), backoff.Infinite))
			return consumer.ConsumeSuccess, nil
		}
		messageEvent.ConvIndex = appendConversationIndexResponse.ConvIndex
		logrus.Infof("[ConvProcess] AppendConversationIndex sucess. index = %v", messageEvent.GetConvIndex())
	}
	receivers := getReceivers(ctx, messageEvent)
	for _, receiver := range receivers {
		err := mq.SendToMq(ctx, "user", strconv.FormatInt(receiver, 10), messageEvent)
		if err != nil {
		}
	}
	return consumer.ConsumeSuccess, nil
	//消费消息(message_parallel_consumer)->校验过滤->重试消息检查消息是否存在(message_api)->获取strategies->callback->处理ext信息
	//->写会话链(inbox_api,V2)->保存消息->(增加thread未读)->写入用户消息队列->与同步MQ互补->失败发送backup队列
}

func getReceivers(ctx context.Context, event *im.MessageEvent) []int64 {
	switch event.GetMsgBody().GetConvType() {
	case int32(im.ConversationType_ConversationType_One_Chat):
		parts := strings.Split(event.GetMsgBody().GetConvId(), ":")
		minId, _ := strconv.ParseInt(parts[0], 10, 64)
		maxId, _ := strconv.ParseInt(parts[1], 10, 64)
		return []int64{minId, maxId}
	case int32(im.ConversationType_ConversationType_Group_Chat):
		//TODO 获取群成员
		return []int64{}
	default:
		return []int64{}
	}
}
