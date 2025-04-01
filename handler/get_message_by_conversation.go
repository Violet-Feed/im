package handler

import (
	"context"
	"github.com/sirupsen/logrus"
	"im/biz"
	"im/proto_gen/common"
	"im/proto_gen/im"
)

func GetMessageByConversation(ctx context.Context, req *im.GetMessageByConversationRequest) (resp *im.GetMessageByConversationResponse) {
	resp = &im.GetMessageByConversationResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	userId := req.GetUserId()
	cores, err := biz.GetConversationCores(ctx, []int64{req.GetConShortId()})
	if err != nil {
		logrus.Errorf("[GetMessageByConversation] GetConversationCores err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return
	}
	if len(cores) == 0 || cores[0].GetStatus() != 0 {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Not_Found_Error, StatusMessage: "会话不存在"}
		return
	}
	//是否群成员
	var isMember int32
	if cores[0].GetConType() == int32(im.ConversationType_One_Chat) {
		isMember = biz.IsSingleMember(ctx, cores[0].GetConId(), userId)
	} else if cores[0].GetConType() == int32(im.ConversationType_Group_Chat) {
		status, err := biz.IsGroupsMember(ctx, []int64{req.GetConShortId()}, userId)
		if err != nil {
			logrus.Errorf("[GetMessageByConversation] IsGroupsMember err. err = %v", err)
			resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
			return
		}
		isMember = status[0]
	}
	if isMember != 1 {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Not_Found_Error, StatusMessage: "非会话成员"}
		return
	}
	//拉取会话链
	msgIds, conIndexs, err := biz.PullConversationIndex(ctx, req.GetConShortId(), req.GetConIndex(), req.GetLimit())
	if err != nil {
		logrus.Errorf("[GetMessageByConversation] PullConversationIndex err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return
	}
	msgBodies, err := biz.GetMessages(ctx, req.GetConShortId(), msgIds)
	if err != nil {
		logrus.Errorf("[GetMessageByConversation] GetMessages err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return
	}
	for i, msgBody := range msgBodies {
		msgBody.ConIndex = conIndexs[i]
	}
	resp.MsgBodies = msgBodies
	return
	//获取core信息(mysql)->获取成员数量(redis+mysql)->判断是否为成员(redis,mysql)->拉取会话链(loadmore)->隐藏撤回消息
}
