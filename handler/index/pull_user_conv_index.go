package index

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"im/dal"
	"im/proto_gen/im"
	"im/util"
	"strconv"
)

func PullUserConvIndex(ctx context.Context, req *im.PullUserConvIndexRequest) (resp *im.PullUserConvIndexResponse, err error) {
	resp = &im.PullUserConvIndexResponse{}
	userId := req.GetUserId()
	userConvIndex := req.GetUserConvIndex()
	limit := req.GetLimit()
	key := fmt.Sprintf("userConvIndex:%d", userId)
	opt := &redis.ZRangeBy{
		Min:   "0",
		Max:   strconv.FormatInt(userConvIndex, 10),
		Count: limit,
	}
	members, err := dal.KvrocksServer.ZRevRangByScoreWithScores(ctx, key, opt)
	if err != nil {
		logrus.Errorf("[PullUserConvIndex] kvrocks ZRevRangByScoreWithScores err. err = %v", err)
		return nil, err
	}
	convShortIds := make([]int64, 0)
	for _, member := range members {
		convShortIds = append(convShortIds, member.Member.(int64))
	}
	resp.ConvShortIds = convShortIds
	count := len(members)
	if count > 0 {
		resp.LastUserConvIndex = util.Int64(int64(members[0].Score))
		resp.NextUserConvIndex = util.Int64(int64(members[count-1].Score) - 1)
	}
	resp.HasMore = util.Bool(count >= int(limit))
	return resp, nil
}
