package consumer

import (
	"context"
	"encoding/json"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/sirupsen/logrus"
	"im/dal/mq"
	"im/handler/conversation"
	"im/handler/index"
	"im/handler/push"
	"im/proto_gen/im"
	"im/util"
	"strconv"
)

func UserProcess(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	var messageEvent *im.MessageEvent
	_ = json.Unmarshal(msgs[0].Body, &messageEvent)
	userId, _ := strconv.ParseInt(msgs[0].GetTags(), 10, 64)
	logrus.Infof("[UserProcess] rocketmq receive message success. message = %v, tag = %v", messageEvent, userId)
	//TODO:判断是否为高频用户(本地+redis,写入高频队列batch)->处理重试消息
	if messageEvent.GetMsgBody().GetMsgType() < 1000 {
		if messageEvent.GetUserConvIndex() == 0 {
			appendUserConvIndexRequest := &im.AppendUserConvIndexRequest{
				UserId:      util.Int64(userId),
				ConvShortId: messageEvent.GetMsgBody().ConvShortId,
			}
			appendUserConvIndexResponse, err := index.AppendUserConvIndex(ctx, appendUserConvIndexRequest)
			if err != nil {
				logrus.Errorf("[UserProcess] AppendUserConvIndex err. err = %v", err)
				retryCount := messageEvent.GetRetryCount()
				if retryCount > 3 {
					logrus.Errorf("[UserProcess] retry too much. message = %v", messageEvent)
				}
				messageEvent.RetryCount = util.Int32(retryCount + 1)
				err = mq.SendMq(ctx, "user", "", messageEvent)
				if err != nil {
					//TODO:无限重试
				}
				return consumer.ConsumeSuccess, nil
			}
			messageEvent.UserConvIndex = appendUserConvIndexResponse.UserConvIndex
			messageEvent.PreUserConvIndex = appendUserConvIndexResponse.PreUserConvIndex
		}
		if messageEvent.GetBadgeCount() == 0 {
			incrConversationBadgeRequest := &im.IncrConversationBadgeRequest{
				UserId:      util.Int64(userId),
				ConvShortId: messageEvent.GetMsgBody().ConvShortId,
			}
			incrConversationBadgeResponse, err := conversation.IncrConversationBadge(ctx, incrConversationBadgeRequest)
			if err != nil {
				logrus.Errorf("[UserProcess] IncrConversationBadge err. err = %v", err)
				retryCount := messageEvent.GetRetryCount()
				if retryCount > 3 {
					logrus.Errorf("[UserProcess] retry too much. message = %v", messageEvent)
				}
				messageEvent.RetryCount = util.Int32(retryCount + 1)
				err = mq.SendMq(ctx, "user", "", messageEvent)
				if err != nil {
					//TODO:无限重试
				}
				return consumer.ConsumeSuccess, nil
			}
			messageEvent.BadgeCount = incrConversationBadgeResponse.BadgeCount
		}
	} else if messageEvent.GetUserCmdIndex() == 0 {
		appendUserCmdIndexRequest := &im.AppendUserCmdIndexRequest{
			UserId: util.Int64(userId),
			MsgId:  messageEvent.GetMsgBody().ConvShortId,
		}
		appendUserCmdIndexResponse, err := index.AppendUserCmdIndex(ctx, appendUserCmdIndexRequest)
		if err != nil {
			logrus.Errorf("[UserProcess] AppendUserCmdIndex err. err = %v", err)
			retryCount := messageEvent.GetRetryCount()
			if retryCount > 3 {
				logrus.Errorf("[UserProcess] retry too much. message = %v", messageEvent)
			}
			messageEvent.RetryCount = util.Int32(retryCount + 1)
			err = mq.SendMq(ctx, "user", "", messageEvent)
			if err != nil {
				//TODO:无限重试
			}
			return consumer.ConsumeSuccess, nil
		}
		messageEvent.UserCmdIndex = appendUserCmdIndexResponse.UserCmdIndex
	}
	//TODO:重试10
	pushRequest := &im.PushRequest{
		MsgBody:          messageEvent.GetMsgBody(),
		ConvIndex:        messageEvent.ConvIndex,
		UserConvIndex:    messageEvent.UserConvIndex,
		PreUserConvIndex: messageEvent.PreUserConvIndex,
		BadgeCount:       messageEvent.BadgeCount,
		UserCmdIndex:     messageEvent.UserCmdIndex,
	}
	_, err := push.Push(ctx, pushRequest)
	if err != nil {
		logrus.Errorf("[UserProcess] Push err. err = %v", err)
	}
	return consumer.ConsumeSuccess, nil

	//消费消息->判断是否为高频用户(本地+redis,写入高频队列batch)->处理重试消息
	//->普通消息:更新最近会话(recent,abase,zset)->增加未读(im_counter_manager_rust,abase,xget获取当前已读数和index,redis对msgid加锁,xset)
	//->命令消息:写入命令链(inbox_api,V2)->处理特殊命令
	//->写入用户链?->push
}
