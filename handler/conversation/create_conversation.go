package conversation

import (
	"context"
	"github.com/sirupsen/logrus"
	"im/handler/conversation/model"
	"im/proto_gen/im"
	"im/util"
)

func CreateConversation(ctx context.Context, req *im.CreateConversationRequest) (resp *im.CreateConversationResponse, err error) {
	resp = &im.CreateConversationResponse{
		BaseResp: &im.BaseResp{StatusCode: im.StatusCode_Success},
	}
	conShortId := util.ConIdGenerator.Generate().Int64()
	//TODO:Identity对conId幂等？
	coreModel := model.PackCoreModel(conShortId, req)
	err = model.InsertCoreInfo(ctx, coreModel)
	if err != nil {
		logrus.Errorf("[CreateConversation] InsertCoreInfo err. err = %v", err)
		resp.BaseResp.StatusCode = im.StatusCode_Server_Error
		return resp, err
	}
	//TODO:发送命令消息，单聊更新setting，群聊添加成员
	if req.GetConType() == int32(im.ConversationType_One_Chat) {
		for _, member := range req.GetMembers() {
			settingModel := model.PackSettingModel(member, conShortId, req)
			err := model.InsertSettingInfo(ctx, settingModel)
			if err != nil {
				logrus.Errorf("[CreateConversation] InsertSettingInfo err. err = %v", err)
				//return?
			}
		}
	}
	if req.GetConType() == int32(im.ConversationType_Group_Chat) {
		addConversationMembersRequest := &im.AddConversationMembersRequest{
			ConShortId: util.Int64(conShortId),
			Members:    req.Members,
			Operator:   req.OwnerId,
		}
		_, err := AddConversationMembers(ctx, addConversationMembersRequest)
		if err != nil {
			logrus.Errorf("[CreateConversation] AddConversationMembers err. err = %v", err)
		}
	}
	coreInfo := model.PackCoreInfo(coreModel)
	resp.ConInfo = coreInfo
	return resp, nil
	//(幂等创建Identity->写redis->)创建/更新core->写redis->发送命令消息，单聊更新setting，群聊添加成员、审核开关
}
