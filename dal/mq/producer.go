package mq

import (
	"context"
	"encoding/json"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/sirupsen/logrus"
	"im/proto_gen/im"
	"im/util/backoff"
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

func SendToMq(ctx context.Context, topic string, key string, message *im.MessageEvent) error {
	body, _ := json.Marshal(message)
	_, err := p.SendSync(ctx, primitive.NewMessage(topic, body).WithShardingKey(key))
	if err != nil {
		logrus.Errorf("[SendMq] mq send message err, err = %v", err)
		return err
	}
	//logrus.Infof("[SendMq] mq send message success, resp = %v", resp)
	return nil
}

func SendToRetry(ctx context.Context, topic string, event *im.MessageEvent) (consumer.ConsumeResult, error) {
	retryCount := event.GetRetryCount()
	if retryCount > 3 {
		logrus.Errorf("[SendToRetry] retry too much. topic = %s, message = %v", topic, event)
		return consumer.ConsumeSuccess, nil
	}
	event.RetryCount = retryCount + 1
	_ = backoff.Retry(func() error {
		return SendToMq(ctx, topic, strconv.FormatInt(event.GetMsgBody().GetConShortId(), 10), event)
	}, backoff.WithMaxRetries(backoff.NewExponentialBackOff(), backoff.Infinite))
	return consumer.ConsumeSuccess, nil
}
