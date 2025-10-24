package consumer

import (
	"context"
	"encoding/json"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/sirupsen/logrus"
	"im/biz"
	"im/biz/constant"
	"im/dal"
	"im/dal/mq"
	"im/proto_gen/im"
	"im/proto_gen/push"
	"im/util/backoff"
	"strconv"
)

func UserProcess(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	var messageEvent *im.MessageEvent
	_ = json.Unmarshal(msgs[0].Body, &messageEvent)
	userId, _ := strconv.ParseInt(msgs[0].GetShardingKey(), 10, 64)
	logrus.Infof("[UserProcess] rocketmq receive message success. message = %v, tag = %v", messageEvent, userId)
	//TODO:判断是否为高频用户(本地+redis,写入高频队列batch)->处理重试消息
	pushRequest := &push.PushRequest{
		UserId: userId,
	}
	if messageEvent.GetMsgBody().GetMsgType() < 100 {
		//写入用户会话链
		if messageEvent.GetUserConIndex() == 0 {
			userConIndex, preUserConIndex, err := biz.AppendUserConIndex(ctx, userId, messageEvent.GetMsgBody().GetConShortId())
			if err != nil {
				logrus.Errorf("[UserProcess] AppendUserConIndex err. err = %v", err)
				return mq.SendToRetry(ctx, constant.IM_USER_TOPIC, messageEvent)
			}
			messageEvent.UserConIndex = userConIndex
			messageEvent.PreUserConIndex = preUserConIndex
		}
		//增加消息总数
		if messageEvent.GetBadgeCount() == 0 {
			if userId != messageEvent.GetMsgBody().GetUserId() && messageEvent.GetMsgBody().GetMsgType() != int32(im.MessageType_Conversation) {
				badgeCount, err := biz.IncrConversationBadge(ctx, userId, messageEvent.GetMsgBody().GetConShortId())
				if err != nil {
					logrus.Errorf("[UserProcess] IncrConversationBadge err. err = %v", err)
					return mq.SendToRetry(ctx, constant.IM_USER_TOPIC, messageEvent)
				}
				messageEvent.BadgeCount = badgeCount
			} else {
				badgeCount, err := biz.GetConversationBadges(ctx, userId, []int64{messageEvent.GetMsgBody().GetConShortId()})
				if err != nil {
					logrus.Errorf("[UserProcess] GetConversationBadges err. err = %v", err)
					return mq.SendToRetry(ctx, constant.IM_USER_TOPIC, messageEvent)
				}
				messageEvent.BadgeCount = badgeCount[0]
			}
		}
		normalPacket := &push.NormalPacket{
			UserConIndex:    messageEvent.GetUserConIndex(),
			PreUserConIndex: messageEvent.GetPreUserConIndex(),
			BadgeCount:      messageEvent.GetBadgeCount(),
			MsgBody:         messageEvent.GetMsgBody(),
		}
		pushRequest.PacketType = push.PacketType_Normal
		pushRequest.NormalPacket = normalPacket
	} else {
		//写入用户命令链
		if messageEvent.GetUserCmdIndex() == 0 {
			userCmdIndex, err := biz.AppendUserCmdIndex(ctx, userId, messageEvent.GetMsgBody().GetMsgId())
			if err != nil {
				logrus.Errorf("[UserProcess] AppendUserCmdIndex err. err = %v", err)
				return mq.SendToRetry(ctx, constant.IM_USER_TOPIC, messageEvent)
			}
			messageEvent.UserCmdIndex = userCmdIndex
		}
		commandPacket := &push.CommandPacket{
			UserCmdIndex: messageEvent.GetUserCmdIndex(),
			MsgBody:      messageEvent.GetMsgBody(),
		}
		pushRequest.PacketType = push.PacketType_Command
		pushRequest.CommandPacket = commandPacket
	}
	err := backoff.Retry(func() error {
		err := dal.PushServer.Push(ctx, pushRequest)
		if err != nil {
			logrus.Errorf("[UserProcess] Push err. err = %v", err)
		}
		return err
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 3))
	if err != nil {
		logrus.Errorf("[UserProcess] Push all err. err = %v", err)
		return mq.SendToRetry(ctx, constant.IM_USER_TOPIC, messageEvent)
	}
	return consumer.ConsumeSuccess, nil
	//消费消息->判断是否为高频用户(本地+redis,写入高频队列batch)->处理重试消息
	//->普通消息:更新最近会话(recent,abase,zset)->增加未读(im_counter_manager_rust,abase,xget获取当前已读数和index,redis对msgid加锁,xset)
	//->命令消息:写入命令链(inbox_api,V2)->处理特殊命令
	//->写入用户链?->push
}
