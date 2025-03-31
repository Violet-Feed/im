package handler

import (
	"github.com/gin-gonic/gin"
	"im/biz"
	"im/proto_gen/common"
	"im/proto_gen/im"
	"im/util"
	"net/http"
)

func checkMessageSendRequest(req *im.MessageSendRequest) bool {
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

func Send(c *gin.Context) {
	resp := &im.MessageSendResponse{}
	var req *im.MessageSendRequest
	err := c.ShouldBindJSON(&req)
	if err != nil || !checkMessageSendRequest(req) {
		c.JSON(http.StatusOK, HttpResponse{
			Code:    common.StatusCode_Param_Error,
			Message: "param error",
			Data:    resp,
		})
		return
	}
	userIdStr, _ := c.Get("userId")
	userId := userIdStr.(int64)
	sendMessageRequest := &im.SendMessageRequest{
		UserId:      util.Int64(userId),
		ConShortId:  req.ConShortId,
		ConId:       req.ConId,
		ConType:     req.ConType,
		ClientMsgId: req.ClientMsgId,
		MsgType:     req.MsgType,
		MsgContent:  req.MsgContent,
	}
	sendMessageResponse, err := biz.SendMessage(c, sendMessageRequest)
	if err != nil {
		c.JSON(http.StatusOK, HttpResponse{
			Code:    sendMessageResponse.GetBaseResp().GetStatusCode(),
			Message: sendMessageResponse.GetBaseResp().GetStatusMessage(),
			Data:    resp,
		})
		return
	}
	c.JSON(http.StatusOK, HttpResponse{
		Code:    common.StatusCode_Success,
		Message: "success",
		Data:    resp,
	})
	return
	//检查合法->生成serverMsgId->异步/同步发送消息
	//检查会话(单聊->创建会话(im_conversation_api)，获取会话coreInfo(im_conversation_api，redis+mysql))->检查消息(integration_callback)
	//->处理@消息->发送消息(message_api)->设置最新用户(im_lastuser_api,abase)
	//发送消息(message_api)->参数校验->(同步)redis加锁消息去重->存入数据库(abase/daas)->判断命令消息->写入消息队列->redis解锁
}
