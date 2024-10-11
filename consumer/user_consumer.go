package consumer

import (
	"context"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/sirupsen/logrus"
)

func UserProcess(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
	logrus.Infof("[UserProcess] rocketmq receive message success. message = %v", msgs[0].Body)

	return consumer.ConsumeSuccess, nil
}
