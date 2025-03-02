package consumer

import (
	"context"
	"encoding/json"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/sirupsen/logrus"
	"im/biz"
	"im/dal/mq"
	"im/proto_gen/im"
	"im/util"
	"im/util/backoff"
	"strconv"
)

func UserProcess(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	var messageEvent *im.MessageEvent
	_ = json.Unmarshal(msgs[0].Body, &messageEvent)
	userId, _ := strconv.ParseInt(msgs[0].GetShardingKey(), 10, 64)
	logrus.Infof("[UserProcess] rocketmq receive message success. message = %v, tag = %v", messageEvent, userId)
	//TODO:判断是否为高频用户(本地+redis,写入高频队列batch)->处理重试消息
	if messageEvent.GetMsgBody().GetMsgType() < 1000 {
		if messageEvent.GetUserConIndex() == 0 {
			userConIndex, preUserConIndex, err := biz.AppendUserConIndex(ctx, userId, messageEvent.GetMsgBody().GetConShortId())
			if err != nil {
				logrus.Errorf("[UserProcess] AppendUserConIndex err. err = %v", err)
				return mq.SendToRetry(ctx, "user", messageEvent)
			}
			messageEvent.UserConIndex = util.Int64(userConIndex)
			messageEvent.PreUserConIndex = util.Int64(preUserConIndex)
		}
		if messageEvent.GetBadgeCount() == 0 {
			badgeCount, err := biz.IncrConversationBadge(ctx, userId, messageEvent.GetMsgBody().GetConShortId())
			if err != nil {
				logrus.Errorf("[UserProcess] IncrConversationBadge err. err = %v", err)
				return mq.SendToRetry(ctx, "user", messageEvent)
			}
			messageEvent.BadgeCount = util.Int64(badgeCount)
		}
	} else if messageEvent.GetUserCmdIndex() == 0 {
		userCmdIndex, err := biz.AppendUserCmdIndex(ctx, userId, messageEvent.GetMsgBody().GetMsgId())
		if err != nil {
			logrus.Errorf("[UserProcess] AppendUserCmdIndex err. err = %v", err)
			return mq.SendToRetry(ctx, "user", messageEvent)
		}
		messageEvent.UserCmdIndex = util.Int64(userCmdIndex)
	}
	pushRequest := &im.PushRequest{
		MsgBody:         messageEvent.GetMsgBody(),
		ReceiverId:      util.Int64(userId),
		ConIndex:        messageEvent.ConIndex,
		UserConIndex:    messageEvent.UserConIndex,
		PreUserConIndex: messageEvent.PreUserConIndex,
		BadgeCount:      messageEvent.BadgeCount,
		UserCmdIndex:    messageEvent.UserCmdIndex,
	}
	err := backoff.Retry(func() error {
		_, err := biz.Push(ctx, pushRequest)
		if err != nil {
			logrus.Errorf("[UserProcess] Push err. err = %v", err)
		}
		return err
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 10))
	if err != nil {
		logrus.Errorf("[UserProcess] Push all err. err = %v", err)
		return mq.SendToRetry(ctx, "user", messageEvent)
	}
	return consumer.ConsumeSuccess, nil
	//消费消息->判断是否为高频用户(本地+redis,写入高频队列batch)->处理重试消息
	//->普通消息:更新最近会话(recent,abase,zset)->增加未读(im_counter_manager_rust,abase,xget获取当前已读数和index,redis对msgid加锁,xset)
	//->命令消息:写入命令链(inbox_api,V2)->处理特殊命令
	//->写入用户链?->push
}
