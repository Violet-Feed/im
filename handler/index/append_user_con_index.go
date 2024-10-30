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

const ConLimit = 1000
const SleepTime = 5 * time.Millisecond

func AppendUserConIndex(ctx context.Context, req *im.AppendUserConIndexRequest) (resp *im.AppendUserConIndexResponse, err error) {
	resp = &im.AppendUserConIndexResponse{}
	userId := req.GetUserId()
	conShortId := req.GetConShortId()
	key := fmt.Sprintf("userConIndex:%d", userId)
	//TODO:redis锁重试3
	//dal.RedisServer.Lock()
	//defer dal.RedisServer.Unlock()
	lastIndex, err := dal.KvrocksServer.ZRangeWithScores(ctx, key, -1, -1)
	if err != nil {
		logrus.Errorf("[AppendUserConIndex] kvrocks ZRangeWithScores err. err = %v", err)
		return nil, err
	}
	var preUserConIndex float64
	if len(lastIndex) == 0 {
		preUserConIndex = 0
	} else {
		preUserConIndex = lastIndex[0].Score
	}
	userConIndex := preUserConIndex + 1
	_, err = dal.KvrocksServer.ZAdd(ctx, key, []redis.Z{
		{
			Member: conShortId,
			Score:  userConIndex,
		},
	})
	if err != nil {
		logrus.Errorf("[AppendUserConIndex] kvrocks ZAdd err. err = %v", err)
		return nil, err
	}
	go dal.KvrocksServer.ZRemRangeByRank(ctx, key, 0, -ConLimit)
	resp.UserConIndex = util.Int64(int64(userConIndex))
	resp.PreUserConIndex = util.Int64(int64(preUserConIndex))
	return resp, nil
}
