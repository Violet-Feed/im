package consumer

import (
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"github.com/sirupsen/logrus"
)

func InitConsumer() {
	rlog.SetLogLevel("warn")
	go func() {
		c, _ := rocketmq.NewPushConsumer(
			consumer.WithNameServer([]string{"127.0.0.1:9876"}),
			consumer.WithGroupName("conversation"),
		)
		if err := c.Subscribe("conversation", consumer.MessageSelector{}, ConvProcess); err != nil {
			logrus.Errorf("[initRocketMq] rocketmq consume conv message err. err = %v", err)
		}
		if err := c.Start(); err != nil {
			logrus.Fatalf("[initRocketMq] rocketmq conv consumer run err. err = %v", err)
		}
	}()
	go func() {
		c, _ := rocketmq.NewPushConsumer(
			consumer.WithNameServer([]string{"127.0.0.1:9876"}),
			consumer.WithGroupName("user"),
		)
		if err := c.Subscribe("user", consumer.MessageSelector{}, UserProcess); err != nil {
			logrus.Errorf("[initRocketMq] rocketmq consume user message err. err = %v", err)
		}
		if err := c.Start(); err != nil {
			logrus.Fatalf("[initRocketMq] rocketmq user consumer run err. err = %v", err)
		}
	}()
}
