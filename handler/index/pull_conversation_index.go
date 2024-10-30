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

func PullConversationIndex(ctx context.Context, req *im.PullConversationIndexRequest) (resp *im.PullConversationIndexResponse, err error) {
	resp = &im.PullConversationIndexResponse{}
	conShortId := req.GetConShortId()
	conIndex := req.GetConIndex()
	limit := req.GetLimit()
	segKey := fmt.Sprintf("convSeg:%d", conShortId)
	seg, err := dal.KvrocksServer.Get(ctx, segKey)
	if errors.Is(err, redis.Nil) {
		return resp, nil
	} else if err != nil {
		logrus.Errorf("[PullConversationIndex] kvrocks get seg err. err = %v", err)
		return nil, err
	}
	indexKey := fmt.Sprintf("convIndex:%d:%s", conShortId, seg)
	length, err := dal.KvrocksServer.LLen(ctx, indexKey)
	if err != nil {
		logrus.Errorf("[PullConversationIndex] kvrocks llen 1 err. err = %v", err)
		return nil, err
	}
	segment, _ := strconv.ParseInt(seg, 10, 64)
	maxIndex := segment*SegmentLimit + length - 1
	if conIndex > maxIndex {
		conIndex = maxIndex
	} else {
		segment = conIndex / SegmentLimit
	}
	messageIds := make([]int64, 0)
	for limit > 0 && segment >= 0 {
		indexKey = fmt.Sprintf("convIndex:%d:%d", conShortId, segment)
		length, err = dal.KvrocksServer.LLen(ctx, indexKey)
		if err != nil {
			logrus.Errorf("[PullConversationIndex] kvrocks llen 2 err. err = %v", err)
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
			logrus.Errorf("[PullConversationIndex] kvrocks lrange err. err = %v", err)
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
	resp.LastConIndex = util.Int64(conIndex)
	return resp, nil
}
