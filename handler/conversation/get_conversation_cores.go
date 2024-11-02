package conversation

import (
	"context"
	"github.com/sirupsen/logrus"
	"im/handler/conversation/model"
	"im/proto_gen/im"
)

func GetConversationCores(ctx context.Context, req *im.GetConversationCoresRequest) (resp *im.GetConversationCoresResponse, err error) {
	resp = &im.GetConversationCoresResponse{
		BaseResp: &im.BaseResp{StatusCode: im.StatusCode_Success},
	}
	coresMap, err := model.GetCoreInfos(ctx, req.GetConShortIds())
	if err != nil {
		logrus.Errorf("[GetConversationCores] GetCoreInfos err. err = %v", err)
		resp.BaseResp.StatusCode = im.StatusCode_Server_Error
		return resp, err
	}
	var coreInfos []*im.ConversationCoreInfo
	for _, id := range req.GetConShortIds() {
		coreInfos = append(coreInfos, model.PackCoreInfo(coresMap[id]))
	}
	resp.CoreInfos = coreInfos
	//TODO:所有群聊成员数量，badge？
	return resp, nil
	//convIds
	//redis mget key:convId,mysql,redis一天
	//并发获取成员数量：useCache:本地缓存key:convId,get,redis get;nil or noUse:redis key:convId,zcard;nil mysql 设置string一分钟,zset永久
}
