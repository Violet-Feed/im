package mq

import (
	"context"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/golang/protobuf/proto"
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

func SendMq(ctx context.Context, message *im.MessageEvent) error {
	body, _ := proto.Marshal(message)
	resp, err := p.SendSync(ctx, primitive.NewMessage("conversation", body).WithShardingKey(strconv.FormatInt(message.GetConvShortId(), 10)))
	if err != nil {
		logrus.Errorf("[SendMq] mq send message err, err = %v", err)
		return err
	}
	logrus.Infof("[SendMq] mq send message success, resp = %v", resp)
	return nil
}
