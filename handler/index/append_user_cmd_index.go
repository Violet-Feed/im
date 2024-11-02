package index

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
)

func AppendUserCmdIndex(ctx context.Context, req *im.AppendUserCmdIndexRequest) (resp *im.AppendUserCmdIndexResponse, err error) {
	resp = &im.AppendUserCmdIndexResponse{
		BaseResp: &im.BaseResp{StatusCode: im.StatusCode_Success},
	}
	userId := req.GetUserId()
	messageId := req.GetMsgId()
	segKey := fmt.Sprintf("user_segment:%d", userId)
	for i := 0; i < 3; i++ {
		seg, err := dal.KvrocksServer.Get(ctx, segKey)
		if errors.Is(err, redis.Nil) {
			opt, err := dal.KvrocksServer.SetNX(ctx, segKey, "0")
			if err != nil {
				logrus.Errorf("[AppendUserCmdIndex] kvrocks SetNX err. err = %v", err)
				resp.BaseResp.StatusCode = im.StatusCode_Server_Error
				return nil, err
			}
			if opt {
				seg = "0"
			} else {
				seg, err = dal.KvrocksServer.Get(ctx, segKey)
				if err != nil {
					logrus.Errorf("[AppendUserCmdIndex] kvrocks Get err. err = %v", err)
					resp.BaseResp.StatusCode = im.StatusCode_Server_Error
					return nil, err
				}
			}
		} else if err != nil {
			logrus.Errorf("[AppendUserCmdIndex] kvrocks Get err. err = %v", err)
			resp.BaseResp.StatusCode = im.StatusCode_Server_Error
			return nil, err
		}
		indexKey := fmt.Sprintf("user_cmd_index:%d:%s", userId, seg)
		subIndex, err := dal.KvrocksServer.RPush(ctx, indexKey, []string{strconv.FormatInt(messageId, 10)})
		if err != nil {
			logrus.Errorf("[AppendUserCmdIndex] kvrocks RPush err. err = %v", err)
			resp.BaseResp.StatusCode = im.StatusCode_Server_Error
			return nil, err
		}
		segment, _ := strconv.ParseInt(seg, 10, 64)
		if subIndex > SegmentLimit {
			newSeg := strconv.FormatInt(segment+1, 10)
			opt, err := dal.KvrocksServer.Cas(ctx, segKey, seg, newSeg)
			if err != nil {
				logrus.Errorf("[AppendUserCmdIndex] kvrocks Cas seg err. err = %v", err)
				resp.BaseResp.StatusCode = im.StatusCode_Server_Error
				return nil, err
			}
			if opt == 1 {
				err = dal.KvrocksServer.Expire(ctx, indexKey, SegmentTTL)
				if err != nil {
					logrus.Errorf("[AppendUserCmdIndex] kvrocks Expire err. err = %v", err)
				}
			}
		} else {
			resp.UserCmdIndex = util.Int64(segment*SegmentLimit + subIndex)
			return resp, nil
		}
	}
	err = errors.New("[AppendUserCmdIndex] err")
	resp.BaseResp.StatusCode = im.StatusCode_RetryTime_Error
	return nil, err
}
