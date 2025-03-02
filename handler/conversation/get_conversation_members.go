package conversation

import (
	"context"
	"github.com/sirupsen/logrus"
	"im/biz/model"
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
	//TODO:not only
	return resp, nil
}
