package conversation

import (
	"context"
	"github.com/sirupsen/logrus"
	"im/handler/conversation/model"
	"im/proto_gen/im"
	"im/util"
)

func GetConversationMembers(ctx context.Context, req *im.GetConversationMembersRequest) (resp *im.GetConversationMemberResponse, err error) {
	resp = &im.GetConversationMemberResponse{
		BaseResp: &im.BaseResp{StatusCode: im.StatusCode_Success},
	}
	if req.GetOnlyId() {
		userIds, err := model.GetUserIdList(ctx, req.GetConShortId())
		if err != nil {
			logrus.Errorf("[GetConversationMembers] GetUserIdList err. err = %v", err)
			resp.BaseResp.StatusCode = im.StatusCode_Server_Error
			return nil, err
		}
		var userInfos []*im.ConversationUserInfo
		for _, id := range userIds {
			userInfos = append(userInfos, &im.ConversationUserInfo{
				UserId: util.Int64(id),
			})
		}
		resp.UserInfos = userInfos
		return resp, nil
	}
	//TODO
	return resp, nil
}

func packUserInfo(userModels []*model.ConversationUserInfo) []*im.ConversationUserInfo {
	var userInfos []*im.ConversationUserInfo
	for _, model := range userModels {
		userInfos = append(userInfos, &im.ConversationUserInfo{
			ConShortId:     util.Int64(model.ConShortId),
			UserId:         util.Int64(model.UserId),
			Privilege:      util.Int32(model.Privilege),
			NickName:       util.String(model.NickName),
			BlockTimeStamp: util.Int64(model.BlockTimeStamp),
			Operator:       util.Int64(model.Operator),
			CreateTime:     util.Int64(model.CreateTime.Unix()),
			ModifyTime:     util.Int64(model.ModifyTime.Unix()),
			Status:         util.Int32(model.Status),
			Extra:          util.String(model.Extra),
		})
	}
	return userInfos
}
