package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"im/dal"
	"im/proto_gen/common"
	"im/proto_gen/im"
	"im/proto_gen/push"
	"im/util"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func SendNotice(ctx context.Context, req *im.SendNoticeRequest) (resp *im.SendNoticeResponse, err error) {
	//按理应该用mq等方式保证数据一致性
	resp = &im.SendNoticeResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	createTime := time.Now().Unix()
	newNotice := false
	noticePacket := &push.NoticePacket{
		Group:  req.GetGroup(),
		OpType: int32(push.NoticeOpType_Send),
	}
	if req.GetAggField() == "" {
		newNotice = true
	} else {
		aggKey := fmt.Sprintf("notice_agg:%d:%d", req.GetUserId(), req.GetGroup())
		noticeIdStr, err := dal.KvrocksServer.HGet(ctx, aggKey, req.GetAggField())
		if errors.Is(err, redis.Nil) {
		} else if err != nil {
			logrus.Warnf("[SendNotice] kvrocks hget err: %v", err)
		}
		if noticeIdStr == "" {
			newNotice = true
		} else {
			noticeId, _ := strconv.ParseInt(noticeIdStr, 10, 64)
			notice := &im.NoticeAggBody{
				SenderId:   req.GetSenderId(),
				CreateTime: createTime,
				Extra:      "",
			}
			noticeByte, _ := json.Marshal(notice)
			aggListKey := fmt.Sprintf("notice_agg_list:%d", noticeId)
			_, err = dal.KvrocksServer.RPush(ctx, aggListKey, []string{string(noticeByte)})
			if err != nil {
				logrus.Errorf("[SendNotice] kvrocks rpush err: %v", err)
				resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
				return resp, err
			}
		}
	}
	if newNotice {
		noticeId := util.NoticeIdGenerator.Generate().Int64()
		notice := &im.NoticeBody{
			NoticeId:      noticeId,
			NoticeType:    req.GetNoticeType(),
			NoticeContent: req.GetNoticeContent(),
			SenderId:      req.GetSenderId(),
			RefId:         req.GetRefId(),
			CreateTime:    createTime,
		}
		noticePacket.NoticeBody = notice
		noticeByte, _ := json.Marshal(notice)
		err = dal.KvrocksServer.Set(ctx, fmt.Sprintf("notice:%d", noticeId), string(noticeByte))
		if err != nil {
			logrus.Errorf("[SendNotice] kvrocks set err: %v", err)
			resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
			return resp, err
		}
		indexKey := fmt.Sprintf("notice_index:%d:%d", req.GetUserId(), req.GetGroup())
		_, err = dal.KvrocksServer.ZAdd(ctx, indexKey, []redis.Z{
			{
				Member: noticeId,
				Score:  float64(createTime),
			},
		})
		if err != nil {
			logrus.Errorf("[SendNotice] kvrocks zadd err: %v", err)
			resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
			return resp, err
		}
		if req.GetAggField() != "" {
			aggKey := fmt.Sprintf("notice_agg:%d:%d", req.GetUserId(), req.GetGroup())
			err := dal.KvrocksServer.HSet(ctx, aggKey, req.GetAggField(), strconv.FormatInt(noticeId, 10))
			if err != nil {
				logrus.Warnf("[SendNotice] kvrocks hset err: %v", err)
			}
		}
	}
	err = dal.KvrocksServer.Incr(ctx, fmt.Sprintf("notice_count:%d:%d", req.GetUserId(), req.GetGroup()))
	if err != nil {
		logrus.Errorf("[SendNotice] kvrocks incr err: %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	noticePacket.NewNotice = newNotice
	pushRequest := &push.PushRequest{
		PacketType:   push.PacketType_Notice,
		UserId:       req.GetUserId(),
		NoticePacket: noticePacket,
	}
	err = dal.PushServer.Push(ctx, pushRequest)
	if err != nil {
		logrus.Errorf("[SendNotice] push err: %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	return resp, nil
}

func GetNoticeList(ctx context.Context, req *im.GetNoticeListRequest) (resp *im.GetNoticeListResponse, err error) {
	resp = &im.GetNoticeListResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	key := fmt.Sprintf("notice_index:%d:%d", req.GetUserId(), req.GetGroup())
	offset := (req.GetPage() - 1) * 10
	noticeIds, err := dal.KvrocksServer.ZRevRange(ctx, key, offset, offset+9)
	if err != nil {
		logrus.Errorf("[GetNoticeList] kvrocks zrevrange err: %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	notices := make([]*im.NoticeBody, 0)
	if len(noticeIds) == 0 {
		resp.Notices = notices
		return resp, nil
	}
	keys := make([]string, 0)
	for _, id := range noticeIds {
		keys = append(keys, fmt.Sprintf("notice:%s", id))
	}
	noticeStr, err := dal.KvrocksServer.MGet(ctx, keys)
	if err != nil {
		logrus.Errorf("[GetNoticeList] kvrocks mget err: %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	aggListKeys := make([]string, 0, len(noticeIds))
	for _, id := range noticeIds {
		aggListKey := fmt.Sprintf("notice_agg_list:%s", id)
		aggListKeys = append(aggListKeys, aggListKey)
	}
	aggLens, err := dal.KvrocksServer.BatchLLen(ctx, aggListKeys)
	if err != nil {
		logrus.Errorf("[GetNoticeList] kvrocks batch llen err: %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	for i, ns := range noticeStr {
		var notice im.NoticeBody
		if err := json.Unmarshal([]byte(ns), &notice); err == nil {
			notice.AggCount = aggLens[i]
			notices = append(notices, &notice)
		}
	}
	resp.Notices = notices
	return resp, nil
}

func GetNoticeAggList(ctx context.Context, req *im.GetNoticeAggListRequest) (resp *im.GetNoticeAggListResponse, err error) {
	resp = &im.GetNoticeAggListResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	key := fmt.Sprintf("notice_agg_list:%d", req.GetNoticeId())
	offset := (req.GetPage() - 1) * 10
	noticeStr, err := dal.KvrocksServer.LRange(ctx, key, offset, offset+9)
	notices := make([]*im.NoticeAggBody, 0)
	for _, ns := range noticeStr {
		var notice im.NoticeAggBody
		if err := json.Unmarshal([]byte(ns), &notice); err == nil {
			notices = append(notices, &notice)
		}
	}
	resp.Notices = notices
	return resp, nil
}

func GetNoticeCount(ctx context.Context, req *im.GetNoticeCountRequest) (resp *im.GetNoticeCountResponse, err error) {
	resp = &im.GetNoticeCountResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	key := fmt.Sprintf("notice_count:%d:%d", req.GetUserId(), req.GetGroup())
	countStr, err := dal.KvrocksServer.Get(ctx, key)
	if errors.Is(err, redis.Nil) {
	} else if err != nil {
		logrus.Errorf("[GetNoticeCount] kvrocks get err: %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	var count int64
	if countStr == "" {
		count = 0
	} else {
		count, _ = strconv.ParseInt(countStr, 10, 64)
	}
	resp.NoticeCount = count
	return resp, nil
}

func MarkNoticeRead(ctx context.Context, req *im.MarkNoticeReadRequest) (resp *im.MarkNoticeReadResponse, err error) {
	resp = &im.MarkNoticeReadResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	key := fmt.Sprintf("notice_count:%d:%d", req.GetUserId(), req.GetGroup())
	err = dal.KvrocksServer.Set(ctx, key, "0")
	if err != nil {
		logrus.Errorf("[MarkNoticeRead] kvrocks set err: %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	aggKey := fmt.Sprintf("notice_agg:%d:%d", req.GetUserId(), req.GetGroup())
	err = dal.KvrocksServer.Del(ctx, aggKey)
	if err != nil {
		logrus.Errorf("[MarkNoticeRead] kvrocks del err: %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	noticePacket := &push.NoticePacket{
		Group:  req.GetGroup(),
		OpType: int32(push.NoticeOpType_MarkRead),
	}
	pushRequest := &push.PushRequest{
		PacketType:   push.PacketType_Notice,
		UserId:       req.GetUserId(),
		NoticePacket: noticePacket,
	}
	err = dal.PushServer.Push(ctx, pushRequest)
	if err != nil {
		logrus.Errorf("[MarkNoticeRead] push err: %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	return resp, nil
}
