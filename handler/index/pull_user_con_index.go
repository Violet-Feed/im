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

func PullUserConIndex(ctx context.Context, req *im.PullUserConIndexRequest) (resp *im.PullUserConIndexResponse, err error) {
	resp = &im.PullUserConIndexResponse{}
	userId := req.GetUserId()
	userConIndex := req.GetUserConIndex()
	limit := req.GetLimit()
	key := fmt.Sprintf("userConIndex:%d", userId)
	opt := &redis.ZRangeBy{
		Min:   "0",
		Max:   strconv.FormatInt(userConIndex, 10),
		Count: limit,
	}
	members, err := dal.KvrocksServer.ZRevRangByScoreWithScores(ctx, key, opt)
	if err != nil {
		logrus.Errorf("[PullUserConIndex] kvrocks ZRevRangByScoreWithScores err. err = %v", err)
		return nil, err
	}
	convShortIds := make([]int64, 0)
	for _, member := range members {
		convShortIds = append(convShortIds, member.Member.(int64))
	}
	resp.ConShortIds = convShortIds
	count := len(members)
	if count > 0 {
		resp.LastUserConIndex = util.Int64(int64(members[0].Score))
		resp.NextUserConIndex = util.Int64(int64(members[count-1].Score) - 1)
	}
	resp.HasMore = util.Bool(count >= int(limit))
	return resp, nil
}
