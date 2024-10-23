package index

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"im/dal"
	"im/proto_gen/im"
	"im/util"
	"time"
)

const ConvLimit = 1000
const SleepTime = 5 * time.Millisecond

func AppendUserConvIndex(ctx context.Context, req *im.AppendUserConvIndexRequest) (resp *im.AppendUserConvIndexResponse, err error) {
	resp = &im.AppendUserConvIndexResponse{}
	userId := req.GetUserId()
	convShortId := req.GetConvShortId()
	key := fmt.Sprintf("userConvIndex:%d", userId)
	//TODO:redis锁重试3
	//dal.RedisServer.Lock()
	//defer dal.RedisServer.Unlock()
	lastIndex, err := dal.KvrocksServer.ZRangeWithScores(ctx, key, -1, -1)
	if err != nil {
		logrus.Errorf("[AppendUserConvIndex] kvrocks ZRangeWithScores err. err = %v", err)
		return nil, err
	}
	var preUserConvIndex float64
	if len(lastIndex) == 0 {
		preUserConvIndex = 0
	} else {
		preUserConvIndex = lastIndex[0].Score
	}
	userConvIndex := preUserConvIndex + 1
	_, err = dal.KvrocksServer.ZAdd(ctx, key, []redis.Z{
		{
			Member: convShortId,
			Score:  userConvIndex,
		},
	})
	if err != nil {
		logrus.Errorf("[AppendUserConvIndex] kvrocks ZAdd err. err = %v", err)
		return nil, err
	}
	go dal.KvrocksServer.ZRemRangeByRank(ctx, key, 0, -ConvLimit)
	resp.UserConvIndex = util.Int64(int64(userConvIndex))
	resp.PreUserConvIndex = util.Int64(int64(preUserConvIndex))
	return resp, nil
}
