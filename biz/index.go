package biz

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"im/dal"
	"strconv"
	"time"
)

const (
	SegmentLimit = 100
	SegmentTTL   = time.Hour * 24 * 180
	ConLimit     = 1000
	SleepTime    = 5 * time.Millisecond
)

func AppendConversationIndex(ctx context.Context, conShortId int64, msgId int64) (int64, error) {
	segKey := fmt.Sprintf("conv_segment:%d", conShortId)
	for i := 0; i < 3; i++ {
		seg, err := dal.KvrocksServer.Get(ctx, segKey)
		if errors.Is(err, redis.Nil) {
			opt, err := dal.KvrocksServer.SetNX(ctx, segKey, "0")
			if err != nil {
				logrus.Errorf("[AppendConversationIndex] kvrocks SetNX err. err = %v", err)
				return 0, err
			}
			if opt {
				seg = "0"
			} else {
				seg, err = dal.KvrocksServer.Get(ctx, segKey)
				if err != nil {
					logrus.Errorf("[AppendConversationIndex] kvrocks Get err. err = %v", err)
					return 0, err
				}
			}
		} else if err != nil {
			logrus.Errorf("[AppendConversationIndex] kvrocks Get err. err = %v", err)
			return 0, err
		}
		indexKey := fmt.Sprintf("conv_index:%d:%s", conShortId, seg)
		subIndex, err := dal.KvrocksServer.RPush(ctx, indexKey, []string{strconv.FormatInt(msgId, 10)})
		if err != nil {
			logrus.Errorf("[AppendConversationIndex] kvrocks RPush err. err = %v", err)
			return 0, err
		}
		segment, _ := strconv.ParseInt(seg, 10, 64)
		if subIndex > SegmentLimit {
			newSeg := strconv.FormatInt(segment+1, 10)
			opt, err := dal.KvrocksServer.Cas(ctx, segKey, seg, newSeg)
			if err != nil {
				logrus.Errorf("[AppendConversationIndex] kvrocks Cas err. err = %v", err)
				return 0, err
			}
			if opt == 1 {
				err = dal.KvrocksServer.Expire(ctx, indexKey, SegmentTTL)
				if err != nil {
					logrus.Errorf("[AppendConversationIndex] kvrocks Expire err. err = %v", err)
				}
			}
		} else {
			return segment*SegmentLimit + subIndex, nil
		}
	}
	return 0, errors.New("retry too much")
}

func AppendUserCmdIndex(ctx context.Context, userId int64, msgId int64) (int64, error) {
	segKey := fmt.Sprintf("user_segment:%d", userId)
	for i := 0; i < 3; i++ {
		seg, err := dal.KvrocksServer.Get(ctx, segKey)
		if errors.Is(err, redis.Nil) {
			opt, err := dal.KvrocksServer.SetNX(ctx, segKey, "0")
			if err != nil {
				logrus.Errorf("[AppendUserCmdIndex] kvrocks SetNX err. err = %v", err)
				return 0, err
			}
			if opt {
				seg = "0"
			} else {
				seg, err = dal.KvrocksServer.Get(ctx, segKey)
				if err != nil {
					logrus.Errorf("[AppendUserCmdIndex] kvrocks Get err. err = %v", err)
					return 0, err
				}
			}
		} else if err != nil {
			logrus.Errorf("[AppendUserCmdIndex] kvrocks Get err. err = %v", err)
			return 0, err
		}
		indexKey := fmt.Sprintf("user_cmd_index:%d:%s", userId, seg)
		subIndex, err := dal.KvrocksServer.RPush(ctx, indexKey, []string{strconv.FormatInt(msgId, 10)})
		if err != nil {
			logrus.Errorf("[AppendUserCmdIndex] kvrocks RPush err. err = %v", err)
			return 0, err
		}
		segment, _ := strconv.ParseInt(seg, 10, 64)
		if subIndex > SegmentLimit {
			newSeg := strconv.FormatInt(segment+1, 10)
			opt, err := dal.KvrocksServer.Cas(ctx, segKey, seg, newSeg)
			if err != nil {
				logrus.Errorf("[AppendUserCmdIndex] kvrocks Cas seg err. err = %v", err)
				return 0, err
			}
			if opt == 1 {
				err = dal.KvrocksServer.Expire(ctx, indexKey, SegmentTTL)
				if err != nil {
					logrus.Errorf("[AppendUserCmdIndex] kvrocks Expire err. err = %v", err)
				}
			}
		} else {
			return segment*SegmentLimit + subIndex, nil
		}
	}
	return 0, errors.New("retry too much")
}

func AppendUserConIndex(ctx context.Context, userId int64, conShortId int64) (int64, int64, error) {
	key := fmt.Sprintf("user_con_index:%d", userId)
	locked := dal.RedisServer.Lock(ctx, key)
	if !locked {
		logrus.Errorf("[AppendUserConIndex] Lock err.")
		return 0, 0, errors.New("retry too much")
	}
	defer dal.RedisServer.Unlock(ctx, key)
	lastIndex, err := dal.KvrocksServer.ZRangeWithScores(ctx, key, -1, -1)
	if err != nil {
		logrus.Errorf("[AppendUserConIndex] kvrocks ZRangeWithScores err. err = %v", err)
		return 0, 0, err
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
		return 0, 0, err
	}
	go dal.KvrocksServer.ZRemRangeByRank(ctx, key, 0, -ConLimit)
	return int64(userConIndex), int64(preUserConIndex), nil
}

func PullConversationIndex(ctx context.Context, conShortId int64, conIndex int64, limit int64) ([]int64, []int64, error) {
	segKey := fmt.Sprintf("conv_segment:%d", conShortId)
	seg, err := dal.KvrocksServer.Get(ctx, segKey)
	if errors.Is(err, redis.Nil) {
		return nil, nil, nil
	} else if err != nil {
		logrus.Errorf("[PullConversationIndex] kvrocks get seg err. err = %v", err)
		return nil, nil, err
	}
	indexKey := fmt.Sprintf("conv_index:%d:%s", conShortId, seg)
	length, err := dal.KvrocksServer.LLen(ctx, indexKey)
	if err != nil {
		logrus.Errorf("[PullConversationIndex] kvrocks llen 1 err. err = %v", err)
		return nil, nil, err
	}
	segment, _ := strconv.ParseInt(seg, 10, 64)
	maxIndex := segment*SegmentLimit + length - 1
	if conIndex > maxIndex {
		conIndex = maxIndex
	} else {
		segment = conIndex / SegmentLimit
	}
	msgIds, conIndexs := make([]int64, 0), make([]int64, 0)
	//反向拉链
	for limit > 0 && segment >= 0 {
		indexKey = fmt.Sprintf("conv_index:%d:%d", conShortId, segment)
		length, err = dal.KvrocksServer.LLen(ctx, indexKey)
		if err != nil {
			logrus.Errorf("[PullConversationIndex] kvrocks llen 2 err. err = %v", err)
			return nil, nil, err
		}
		if length == 0 && len(msgIds) > 0 {
			return msgIds, conIndexs, nil
		}
		var start, stop int64
		if length > limit {
			start, stop = length-limit, length-1
		} else {
			start, stop = 0, length-1
		}
		subMsgIds, err := dal.KvrocksServer.LRange(ctx, indexKey, start, stop)
		if err != nil {
			logrus.Errorf("[PullConversationIndex] kvrocks lrange err. err = %v", err)
			return nil, nil, err
		}
		for i := 0; i < len(subMsgIds); i++ {
			msgId, _ := strconv.ParseInt(subMsgIds[i], 10, 64)
			msgIds = append(msgIds, msgId)
			conIndexs = append(conIndexs, segment*SegmentLimit+start+int64(i)+1)
		}
		segment--
		limit -= stop - start + 1
	}
	return msgIds, conIndexs, nil
}

func PullUserCmdIndex(ctx context.Context, userId int64, userCmdIndex int64, limit int64) ([]int64, int64, error) {
	segKey := fmt.Sprintf("user_segment:%d", userId)
	seg, err := dal.KvrocksServer.Get(ctx, segKey)
	if errors.Is(err, redis.Nil) {
		return nil, 0, nil
	} else if err != nil {
		logrus.Errorf("[PullUserCmdIndex] kvrocks get seg err. err = %v", err)
		return nil, 0, err
	}
	indexKey := fmt.Sprintf("user_cmd_index:%d:%s", userId, seg)
	length, err := dal.KvrocksServer.LLen(ctx, indexKey)
	if err != nil {
		logrus.Errorf("[PullUserCmdIndex] kvrocks llen 1 err. err = %v", err)
		return nil, 0, err
	}
	segment, _ := strconv.ParseInt(seg, 10, 64)
	maxIndex := segment*SegmentLimit + length
	if userCmdIndex > maxIndex {
		userCmdIndex = maxIndex
	} else {
		segment = userCmdIndex / SegmentLimit
	}
	msgIds := make([]int64, 0)
	for limit > 0 && segment >= 0 {
		indexKey = fmt.Sprintf("user_cmd_index:%d:%d", userId, segment)
		length, err = dal.KvrocksServer.LLen(ctx, indexKey)
		if err != nil {
			logrus.Errorf("[PullUserCmdIndex] kvrocks llen 2 err. err = %v", err)
			return nil, 0, err
		}
		if length == 0 && len(msgIds) > 0 {
			return msgIds, userCmdIndex, nil
		}
		var start, stop int64
		if length > limit {
			start, stop = length-limit, length-1
		} else {
			start, stop = 0, length-1
		}
		subMsgIds, err := dal.KvrocksServer.LRange(ctx, indexKey, start, stop)
		if err != nil {
			logrus.Errorf("[PullUserCmdIndex] kvrocks lrange err. err = %v", err)
			return nil, 0, err
		}
		for i := len(subMsgIds) - 1; i >= 0; i-- {
			msgId, _ := strconv.ParseInt(subMsgIds[i], 10, 64)
			msgIds = append(msgIds, msgId)
		}
		segment--
		limit -= stop - start + 1
	}
	return msgIds, userCmdIndex, nil
}

func PullUserConIndex(ctx context.Context, userId int64, userConIndex int64, limit int64) ([]int64, []int64, bool, error) {
	key := fmt.Sprintf("user_con_index:%d", userId)
	opt := &redis.ZRangeBy{
		Min:   "0",
		Max:   strconv.FormatInt(userConIndex, 10),
		Count: limit + 1,
	}
	members, err := dal.KvrocksServer.ZRevRangByScoreWithScores(ctx, key, opt)
	if err != nil {
		logrus.Errorf("[PullUserConIndex] kvrocks ZRevRangByScoreWithScores err. err = %v", err)
		return nil, nil, false, err
	}
	conShortIds, userConIndexs := make([]int64, 0), make([]int64, 0)
	for _, member := range members {
		conShortId, _ := strconv.ParseInt(member.Member.(string), 10, 64)
		conShortIds = append(conShortIds, conShortId)
		userConIndexs = append(userConIndexs, int64(member.Score))
	}
	count := len(members)
	hasMore := false
	if int64(count) > limit {
		hasMore = true
		conShortIds = conShortIds[:limit]
		userConIndexs = userConIndexs[:limit]
	}
	return conShortIds, userConIndexs, hasMore, nil
}
