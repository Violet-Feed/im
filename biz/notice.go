package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"im/dal"
	"im/proto_gen/common"
	"im/proto_gen/im"
	"im/util"
	"strconv"
	"time"
)

func SendNotice(ctx context.Context, req *im.SendNoticeRequest) (resp *im.SendNoticeResponse, err error) {
	resp = &im.SendNoticeResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	noticeId := util.NoticeIdGenerator.Generate().Int64()
	createTime := time.Now().Unix()
	notice := &im.NoticeBody{
		NoticeId:      noticeId,
		NoticeType:    req.GetNoticeType(),
		NoticeContent: req.GetNoticeContent(),
		CreateTime:    createTime,
	}
	noticeJson, err := json.Marshal(notice)
	if err != nil {
		logrus.Errorf("[SendNotice] json marshal err: %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	err = dal.KvrocksServer.Set(ctx, fmt.Sprintf("notice:%d", noticeId), string(noticeJson))
	if err != nil {
		logrus.Errorf("[SendNotice] kvrocks set err: %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	if req.GetRefId() != 0 {
		refKey := fmt.Sprintf("notice_ref:%d", req.GetUserId())
		refField := fmt.Sprintf("%s:%d", req.GetRefType(), req.GetRefId())
		res, err := dal.KvrocksServer.HGet(ctx, refKey, refField)
		if err != nil {
			logrus.Errorf("[SendNotice] kvrocks hget err: %v", err)
			resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
			return resp, err
		}
		if res != "" {
			aggKey := fmt.Sprintf("notice_agg:%d", noticeId)
			_, err := dal.KvrocksServer.RPush(ctx, aggKey, []string{res})
			if err != nil {
				logrus.Errorf("[SendNotice] kvrocks rpush err: %v", err)
				resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
				return resp, err
			}
		} else {
			err := dal.KvrocksServer.HSet(ctx, refKey, refField, strconv.FormatInt(noticeId, 10))
			if err != nil {
				logrus.Errorf("[SendNotice] kvrocks hset err: %v", err)
				resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
				return resp, err
			}
			key := fmt.Sprintf("notice_index:%d", req.GetUserId())
			_, err = dal.KvrocksServer.ZAdd(ctx, key, []redis.Z{
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
		}
	} else {
		key := fmt.Sprintf("notice_index:%d:%d", req.GetUserId(), req.GetGroup())
		_, err := dal.KvrocksServer.ZAdd(ctx, key, []redis.Z{
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
	}
	err = dal.KvrocksServer.Incr(ctx, fmt.Sprintf("notice_count:%d", req.GetUserId()))
	if err != nil {
		logrus.Errorf("[SendNotice] kvrocks incr err: %v", err)
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
	notices := make([]*im.NoticeBody, 0)
	for _, ns := range noticeStr {
		var notice im.NoticeBody
		err := json.Unmarshal([]byte(ns), &notice)
		if err != nil {
			logrus.Errorf("[GetNoticeList] json unmarshal err: %v", err)
			resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
			return resp, err
		}
		notices = append(notices, &notice)
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
	if err != nil {
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
	refKey := fmt.Sprintf("notice_ref:%d", req.GetUserId())
	err = dal.KvrocksServer.Del(ctx, refKey)
	if err != nil {
		logrus.Errorf("[MarkNoticeRead] kvrocks del err: %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	return resp, nil
}
