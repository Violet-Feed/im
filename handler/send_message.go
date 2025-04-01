package handler

import (
	"context"
	"im/biz"
	"im/proto_gen/common"
	"im/proto_gen/im"
)

func checkMessageSendRequest(req *im.SendMessageRequest) bool {
	if req.GetConId() == "" {
		return false
	}
	if req.GetConType() < 1 || req.GetConType() > 5 {
		return false
	}
	if req.GetClientMsgId() == 0 {
		return false
	}
	if req.GetMsgType() < 1 || req.GetMsgType() > 5 && req.GetMsgType() != 1000 && req.GetMsgType() != 1001 {
		return false
	}
	if req.GetMsgContent() == "" {
		return false
	}
	return true
}

func SendMessage(ctx context.Context, req *im.SendMessageRequest) (resp *im.SendMessageResponse) {
	resp = &im.SendMessageResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	if !checkMessageSendRequest(req) {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Param_Error, StatusMessage: "参数错误"}
		return
	}
	resp, err := biz.SendMessage(ctx, req)
	if err != nil {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return
	}
	return
	//检查合法->生成serverMsgId->异步/同步发送消息
	//检查会话(单聊->创建会话(im_conversation_api)，获取会话coreInfo(im_conversation_api，redis+mysql))->检查消息(integration_callback)
	//->处理@消息->发送消息(message_api)->设置最新用户(im_lastuser_api,abase)
	//发送消息(message_api)->参数校验->(同步)redis加锁消息去重->存入数据库(abase/daas)->判断命令消息->写入消息队列->redis解锁
}
