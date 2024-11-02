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

func PullUserCmdIndex(ctx context.Context, req *im.PullUserCmdIndexRequest) (resp *im.PullUserCmdIndexResponse, err error) {
	resp = &im.PullUserCmdIndexResponse{
		BaseResp: &im.BaseResp{StatusCode: im.StatusCode_Success},
	}
	userId := req.GetUserId()
	userCmdIndex := req.GetUserCmdIndex()
	limit := req.GetLimit()
	segKey := fmt.Sprintf("user_segment:%d", userId)
	seg, err := dal.KvrocksServer.Get(ctx, segKey)
	if errors.Is(err, redis.Nil) {
		return resp, nil
	} else if err != nil {
		logrus.Errorf("[PullUserCmdIndex] kvrocks get seg err. err = %v", err)
		resp.BaseResp.StatusCode = im.StatusCode_Server_Error
		return nil, err
	}
	indexKey := fmt.Sprintf("user_cmd_index:%d:%s", userId, seg)
	length, err := dal.KvrocksServer.LLen(ctx, indexKey)
	if err != nil {
		logrus.Errorf("[PullUserCmdIndex] kvrocks llen 1 err. err = %v", err)
		resp.BaseResp.StatusCode = im.StatusCode_Server_Error
		return nil, err
	}
	segment, _ := strconv.ParseInt(seg, 10, 64)
	maxIndex := segment*SegmentLimit + length - 1
	if userCmdIndex > maxIndex {
		userCmdIndex = maxIndex
	} else {
		segment = userCmdIndex / SegmentLimit
	}
	messageIds := make([]int64, 0)
	for limit > 0 && segment >= 0 {
		indexKey = fmt.Sprintf("user_cmd_index:%d:%d", userId, segment)
		length, err = dal.KvrocksServer.LLen(ctx, indexKey)
		if err != nil {
			logrus.Errorf("[PullUserCmdIndex] kvrocks llen 2 err. err = %v", err)
			resp.BaseResp.StatusCode = im.StatusCode_Server_Error
			return nil, err
		}
		if length == 0 && len(messageIds) > 0 {
			resp.MsgIds = messageIds
			return resp, nil
		}
		var start, stop int64
		if length > limit {
			start, stop = length-limit, length-1
		} else {
			start, stop = 0, length-1
		}
		subMessageIds, err := dal.KvrocksServer.LRange(ctx, indexKey, start, stop)
		if err != nil {
			logrus.Errorf("[PullUserCmdIndex] kvrocks lrange err. err = %v", err)
			resp.BaseResp.StatusCode = im.StatusCode_Server_Error
			return nil, err
		}
		for i := len(subMessageIds) - 1; i >= 0; i-- {
			messageId, _ := strconv.ParseInt(subMessageIds[i], 10, 64)
			messageIds = append(messageIds, messageId)
		}
		segment--
		limit -= stop - start + 1
	}
	resp.MsgIds = messageIds
	resp.LastUserCmdIndex = util.Int64(userCmdIndex)
	return resp, nil
}
