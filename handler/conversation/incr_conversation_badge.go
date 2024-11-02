package conversation

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"im/dal"
	"im/proto_gen/im"
	"im/util"
	"strconv"
	"time"
)

const SleepTime = 5 * time.Millisecond

func IncrConversationBadge(ctx context.Context, req *im.IncrConversationBadgeRequest) (resp *im.IncrConversationBadgeResponse, err error) {
	resp = &im.IncrConversationBadgeResponse{
		BaseResp: &im.BaseResp{StatusCode: im.StatusCode_Success},
	}
	userId := req.GetUserId()
	conShortId := req.GetConShortId()
	//TODO:为什么有userId
	key := fmt.Sprintf("badge:%d:%d", userId, conShortId)
	for i := 0; i < 3; i++ {
		badgeCnt, err := dal.KvrocksServer.Get(ctx, key)
		if errors.Is(err, redis.Nil) {
			newBadgeCntNum := int64(1)
			newBadgeCnt := "1"
			opt, err := dal.KvrocksServer.SetNX(ctx, key, newBadgeCnt)
			if err != nil {
				logrus.Errorf("[IncrConversationBadge] kvrocks SetNX err. err = %v", err)
				resp.BaseResp.StatusCode = im.StatusCode_Server_Error
				return nil, err
			}
			if opt {
				resp.BadgeCount = util.Int64(newBadgeCntNum)
				return resp, nil
			}
		} else if err != nil {
			logrus.Errorf("[IncrConversationBadge] kvrocks Get err. err = %v", err)
			resp.BaseResp.StatusCode = im.StatusCode_Server_Error
			return nil, err
		} else {
			badgeCntNum, _ := strconv.ParseInt(badgeCnt, 10, 64)
			newBadgeCntNum := badgeCntNum + 1
			newBadgeCnt := strconv.FormatInt(newBadgeCntNum, 10)
			opt, err := dal.KvrocksServer.Cas(ctx, key, badgeCnt, newBadgeCnt)
			if err != nil {
				logrus.Errorf("[IncrConversationBadge] kvrocks Cas err. err = %v", err)
				resp.BaseResp.StatusCode = im.StatusCode_Server_Error
				return nil, err
			}
			if opt == 1 {
				resp.BadgeCount = util.Int64(newBadgeCntNum)
				return resp, nil
			}
		}
		time.Sleep(SleepTime)
	}
	return nil, errors.New("[IncrConversationBadge] retry too much")
}
