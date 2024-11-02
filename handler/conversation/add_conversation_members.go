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
		resp.BaseResp.StatusCode = im.StatusCode_OverLimit_Error
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
	//TODO:个人认为流程：redis加锁，获取成员数量判断是否超限，存入数据库，发送普通消息，解锁
	//TODO:入单链，对于新成员设置已读起点终点minIndex，入用户链，对于新成员获取badge设置readBadge
	//获取保存badgeCount
	//保存userModels
	//再次判断limit，删除成员
	//拉会话链，获取index起点,设置已读起点
	//发送进群命令消息
	//发送挡板消息？
	//设置minIndex，已读终点
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
