package mq

import (
	"context"
	"encoding/json"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/sirupsen/logrus"
	"im/proto_gen/im"
	"strconv"
)

var p rocketmq.Producer

func init() {
	var err error
	p, err = rocketmq.NewProducer(producer.WithNameServer([]string{"127.0.0.1:9876"}), producer.WithRetry(1))
	if err != nil {
		logrus.Fatalf("[init] rocketmq producer create err. err = %v", err)
	}
	err = p.Start()
	if err != nil {
		logrus.Fatalf("[init] rocketmq producer run err. err = %v", err)
	}
}

func SendMq(ctx context.Context, topic string, tag string, message *im.MessageEvent) error {
	body, _ := json.Marshal(message)
	_, err := p.SendSync(ctx, primitive.NewMessage(topic, body).
		WithShardingKey(strconv.FormatInt(message.GetMsgBody().GetConvShortId(), 10)).
		WithTag(tag))
	if err != nil {
		logrus.Errorf("[SendMq] mq send message err, err = %v", err)
		return err
	}
	//logrus.Infof("[SendMq] mq send message success, resp = %v", resp)
	return nil
}
