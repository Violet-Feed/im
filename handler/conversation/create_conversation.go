package conversation

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"im/dal"
	"im/dal/mysql"
	"im/proto_gen/im"
	"im/util"
	"time"
)

func CreateConversation(ctx context.Context, req *im.CreateConversationRequest) (resp *im.CreateConversationResponse, err error) {
	resp = &im.CreateConversationResponse{}
	conShortId := util.ConvIdGenerator.Generate().Int64()
	resp.ConvShortId = util.Int64(conShortId)
	core := packCoreInfo(conShortId, req)
	err = mysql.InsertCoreInfo(core)
	if err != nil {
		logrus.Errorf("[CreateConversation] mysql insert core err. err = %v", err)
		return nil, err
	}
	coreByte, err := json.Marshal(core)
	if err == nil {
		key := fmt.Sprintf("core:%d", conShortId)
		_ = dal.RedisServer.Set(ctx, key, string(coreByte), 1*time.Minute)
	}
	//TODO:发送命令消息，单聊更新setting，群聊添加成员

	return resp, nil
	//(幂等创建Identity->写redis->)创建/更新core->写redis->发送命令消息，单聊更新setting，群聊添加成员、审核开关
}

func packCoreInfo(conShortId int64, req *im.CreateConversationRequest) *mysql.ConversationCoreInfo {
	core := &mysql.ConversationCoreInfo{
		ConShortId:  conShortId,
		ConId:       req.GetConvId(),
		ConType:     req.GetConvType(),
		Name:        req.GetName(),
		AvatarUri:   req.GetAvatarUri(),
		Description: req.GetDescription(),
		Notice:      req.GetNotice(),
		OwnerId:     req.GetOwnerId(),
		Extra:       req.GetExtra(),
	}
	//TODO:获取req里的time？
	curTime := time.Now()
	core.CreateTime = curTime
	core.ModifyTime = curTime
	return core
}
