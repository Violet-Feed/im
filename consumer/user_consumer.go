package consumer

import (
	"context"
	"encoding/json"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/sirupsen/logrus"
	"im/proto_gen/im"
)

func UserProcess(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	var messageEvent *im.MessageEvent
	_ = json.Unmarshal(msgs[0].Body, &messageEvent)
	userId := msgs[0].GetTags()
	logrus.Infof("[UserProcess] rocketmq receive message success. message = %v, tag = %v", messageEvent, userId)
	//消费消息->判断是否为高频用户(本地+redis,写入高频队列batch)->处理重试消息->处理普通消息->更新最近会话(recent,abase,zset)
	//->增加未读(im_counter_manager_rust,abase,xget获取当前已读数和index,redis对msgid加锁,xset)->写入用户链->处理命令消息
	//->写入命令链(inbox_api,V2)->处理特殊命令->写入用户链->push
	return consumer.ConsumeSuccess, nil
}
