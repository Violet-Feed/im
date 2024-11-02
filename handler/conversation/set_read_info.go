package conversation

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"im/dal"
	"im/proto_gen/im"
	"strconv"
)

func SetReadInfo(ctx context.Context, req *im.SetReadInfoRequest) (resp *im.SetReadInfoResponse, err error) {
	resp = &im.SetReadInfoResponse{
		BaseResp: &im.BaseResp{StatusCode: im.StatusCode_Success},
	}
	userId := req.GetUserId()
	conShortId := req.GetConShortId()
	readIndexStart := req.GetReadIndexStart()
	readIndexEnd := req.GetReadIndexEnd()
	readBadgeCount := req.GetReadBadgeCount()
	if readIndexStart != 0 {
		readIndexStartKey := fmt.Sprintf("read_index_start:%d:%d", userId, conShortId)
		err := dal.KvrocksServer.Set(ctx, readIndexStartKey, strconv.FormatInt(readIndexStart, 10))
		if err != nil {
			logrus.Errorf("[SetReadInfo] set readIndexStart err. err = %v", err)
			resp.BaseResp.StatusCode = im.StatusCode_Server_Error
			return resp, err
		}
	}
	readIndexEndKey := fmt.Sprintf("read_index_end:%d:%d", userId, conShortId)
	err = dal.KvrocksServer.Set(ctx, readIndexEndKey, strconv.FormatInt(readIndexEnd, 10))
	if err != nil {
		logrus.Errorf("[SetReadInfo] set readIndexEndKey err. err = %v", err)
		resp.BaseResp.StatusCode = im.StatusCode_Server_Error
		return resp, err
	}
	readBadgeCountKey := fmt.Sprintf("read_badge_count:%d:%d", userId, conShortId)
	err = dal.KvrocksServer.Set(ctx, readBadgeCountKey, strconv.FormatInt(readBadgeCount, 10))
	if err != nil {
		logrus.Errorf("[SetReadInfo] set readBadgeCountKey err. err = %v", err)
		resp.BaseResp.StatusCode = im.StatusCode_Server_Error
		return resp, err
	}
	return resp, nil
}
