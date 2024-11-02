package conversation

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"im/dal"
	"im/proto_gen/im"
	"strconv"
)

func GetConversationBadges(ctx context.Context, req *im.GetConversationBadgesRequest) (resp *im.GetConversationBadgesResponse, err error) {
	resp = &im.GetConversationBadgesResponse{
		BaseResp: &im.BaseResp{StatusCode: im.StatusCode_Success},
	}
	userId := req.GetUserId()
	conShortIds := req.GetConShortIds()
	keys := make([]string, 0)
	for _, id := range conShortIds {
		key := fmt.Sprintf("badge:%d:%d", userId, id)
		keys = append(keys, key)
	}
	countStrs, err := dal.KvrocksServer.MGet(ctx, keys)
	if err != nil {
		logrus.Errorf("[GetConversationBadge] kvrocks MGet err. err = %v", err)
		resp.BaseResp.StatusCode = im.StatusCode_Server_Error
		return nil, err
	}
	counts := make([]int64, 0)
	for _, countStr := range countStrs {
		if countStr == "" {
			counts = append(counts, 0)
		} else {
			count, _ := strconv.ParseInt(countStr, 10, 64)
			counts = append(counts, count)
		}
	}
	resp.BadgeCounts = counts
	return resp, nil
}
