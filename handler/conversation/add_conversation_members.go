package conversation

import (
	"context"
	"github.com/sirupsen/logrus"
	"im/handler/conversation/model"
	"im/proto_gen/im"
	"time"
)

const ConversationLimit = 100

func AddConversationMembers(ctx context.Context, req *im.AddConversationMembersRequest) (resp *im.AddConversationMembersResponse, err error) {
	resp = &im.AddConversationMembersResponse{
		BaseResp: &im.BaseResp{StatusCode: im.StatusCode_Success},
	}
	conShortId := req.GetConShortId()
	//TODO:群是否存在
	count, err := model.GetUserCount(ctx, conShortId)
	if err != nil {
		logrus.Errorf("[AddConversationMembers] GetUserCount err. err = %v", err)
		resp.BaseResp.StatusCode = im.StatusCode_Server_Error
		return resp, err
	}
	if count+len(req.GetMembers()) > ConversationLimit {
		resp.BaseResp.StatusCode = im.StatusCode_Limit_Error
		return resp, nil
	}
	var userModels []*model.ConversationUserInfo
	for _, member := range req.GetMembers() {
		userModel := packUserModel(member, req)
		if count == 0 && member == req.GetOperator() {
			userModel.Privilege = 0
		}
		userModels = append(userModels, userModel)
	}
	return resp, nil
}

func packUserModel(userId int64, req *im.AddConversationMembersRequest) *model.ConversationUserInfo {
	user := &model.ConversationUserInfo{
		ConShortId: req.GetConShortId(),
		UserId:     userId,
		Privilege:  1,
		Operator:   req.GetOperator(),
	}
	curTime := time.Now()
	user.CreateTime = curTime
	user.ModifyTime = curTime
	return user
}
