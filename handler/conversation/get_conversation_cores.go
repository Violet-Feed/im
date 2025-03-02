package conversation

import (
	"context"
	"github.com/sirupsen/logrus"
	model2 "im/biz/model"
	"im/proto_gen/im"
	"im/util"
	"sync"
)

func GetConversationCores(ctx context.Context, req *im.GetConversationCoresRequest) (resp *im.GetConversationCoresResponse, err error) {
	resp = &im.GetConversationCoresResponse{
		BaseResp: &im.BaseResp{StatusCode: im.StatusCode_Success},
	}
	coresMap, err := model2.GetCoreInfos(ctx, req.GetConShortIds())
	if err != nil {
		logrus.Errorf("[GetConversationCores] GetCoreInfos err. err = %v", err)
		resp.BaseResp.StatusCode = im.StatusCode_Server_Error
		return resp, err
	}
	var coreInfos []*im.ConversationCoreInfo
	for _, id := range req.GetConShortIds() {
		coreInfos = append(coreInfos, model2.PackCoreInfo(coresMap[id]))
	}
	//TODO:所有群聊成员数量，badge？
	wg := sync.WaitGroup{}
	badgeChan := make([]chan int, len(req.GetConShortIds()))
	for i, conShortId := range req.GetConShortIds() {
		wg.Add(1)
		go func(i int, conShortId int64) {
			defer wg.Done()
			count, err := model2.GetUserCount(ctx, conShortId)
			if err != nil {
				logrus.Errorf("[GetConversationCores] GetUserCount err. err = %v", err)
			}
			badgeChan[i] <- count
		}(i, conShortId)
	}
	wg.Wait()
	for i := 0; i < len(req.GetConShortIds()); i++ {
		coreInfos[i].MemberCount = util.Int32(int32(<-badgeChan[i]))
	}
	resp.CoreInfos = coreInfos
	return resp, nil
	//redis mget key:convId,mysql,redis一天
	//并发获取成员数量：useCache:本地缓存key:convId,get,redis get;nil or noUse:redis key:convId,zcard;nil mysql 设置string一分钟,zset永久
}
