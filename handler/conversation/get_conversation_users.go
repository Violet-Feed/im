package conversation

import (
	"context"
	"github.com/sirupsen/logrus"
	"im/handler/conversation/model"
	"im/proto_gen/im"
)

func GetConversationUsers(ctx context.Context, conShortId int64, userIds []int64) ([]*im.ConversationUserInfo, error) {
	userMap, err := model.GetUserInfos(ctx, conShortId, userIds, true)
	if err != nil {
		logrus.Errorf("[GetConversationMembers] GetUserInfos err. err = %v", err)
		return nil, err
	}
	var userInfos []*im.ConversationUserInfo
	for _, id := range userIds {
		userInfos = append(userInfos, model.PackUserInfo(userMap[id]))
	}
	return userInfos, nil
}
