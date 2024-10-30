package conversation

import (
	"context"
	"github.com/sirupsen/logrus"
	"im/handler/conversation/model"
	"im/proto_gen/im"
	"im/util"
	"strconv"
	"time"
)

func CreateConversation(ctx context.Context, req *im.CreateConversationRequest) (resp *im.CreateConversationResponse, err error) {
	resp = &im.CreateConversationResponse{
		BaseResp: &im.BaseResp{StatusCode: im.StatusCode_Success},
	}
	conShortId := util.ConIdGenerator.Generate().Int64()
	//TODO:Identity对conId幂等？
	coreModel := packCoreModel(conShortId, req)
	err = model.InsertCoreInfo(ctx, coreModel)
	if err != nil {
		logrus.Errorf("[CreateConversation] InsertCoreInfo err. err = %v", err)
		resp.BaseResp.StatusCode = im.StatusCode_Server_Error
		return resp, err
	}
	//TODO:发送命令消息，单聊更新setting，群聊添加成员
	if req.GetConType() == int32(im.ConversationType_One_Chat) {
		for _, member := range req.GetMembers() {
			settingModel := packSettingModel(member, conShortId, req)
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
	coreInfo := packCoreInfo(coreModel)
	resp.ConInfo = coreInfo
	return resp, nil
	//(幂等创建Identity->写redis->)创建/更新core->写redis->发送命令消息，单聊更新setting，群聊添加成员、审核开关
}

func packCoreModel(conShortId int64, req *im.CreateConversationRequest) *model.ConversationCoreInfo {
	core := &model.ConversationCoreInfo{
		ConShortId:  conShortId,
		ConId:       req.GetConId(),
		ConType:     req.GetConType(),
		Name:        req.GetName(),
		AvatarUri:   req.GetAvatarUri(),
		Description: req.GetDescription(),
		Notice:      req.GetNotice(),
		OwnerId:     req.GetOwnerId(),
		Extra:       req.GetExtra(),
	}
	if req.GetConType() == int32(im.ConversationType_Group_Chat) {
		core.ConId = strconv.FormatInt(conShortId, 10)
	}
	//TODO:获取req里的time？
	curTime := time.Now()
	core.CreateTime = curTime
	core.ModifyTime = curTime
	return core
}

func packCoreInfo(model *model.ConversationCoreInfo) *im.ConversationCoreInfo {
	core := &im.ConversationCoreInfo{
		ConShortId:  util.Int64(model.ConShortId),
		ConId:       util.String(model.ConId),
		ConType:     util.Int32(model.ConType),
		Name:        util.String(model.Name),
		AvatarUri:   util.String(model.AvatarUri),
		Description: util.String(model.Description),
		Notice:      util.String(model.Notice),
		OwnerId:     util.Int64(model.OwnerId),
		CreateTime:  util.Int64(model.CreateTime.Unix()),
		ModifyTime:  util.Int64(model.ModifyTime.Unix()),
		Status:      util.Int32(model.Status),
		Extra:       util.String(model.Extra),
	}
	return core
}

func packSettingModel(userId int64, conShortId int64, req *im.CreateConversationRequest) *model.ConversationSettingInfo {
	setting := &model.ConversationSettingInfo{
		UserId:     userId,
		ConShortId: conShortId,
		ConId:      req.GetConId(),
		ConType:    req.GetConType(),
		Extra:      req.GetExtra(),
	}
	curTime := time.Now()
	setting.ModifyTime = curTime
	return setting
}
