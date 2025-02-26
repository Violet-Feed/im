package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"im/handler/message"
	"im/proto_gen/im"
	"im/util"
	"net/http"
)

func checkMessageSendRequest(c *gin.Context, req *im.MessageSendRequest) bool {
	//TODO：参数校验
	if req.GetConId() == "" {
		logrus.Info("ConId is empty")
		return false
	}
	if req.GetConType() < 1 || req.GetConType() > 5 {
		logrus.Info("ConType is invalid")
		return false
	}
	if req.GetMsgType() < 1 || req.GetMsgType() > 5 && req.GetMsgType() != 1000 {
		logrus.Info("MsgType is invalid")
		return false
	}
	if req.GetMsgContent() == "" {
		logrus.Info("MsgContent is empty")
		return false
	}
	return true
}

func Send(c *gin.Context) {
	var req *im.MessageSendRequest
	err := c.ShouldBindJSON(&req)
	if err != nil || !checkMessageSendRequest(c, req) {
		c.JSON(http.StatusOK, im.StatusCode_Param_Error)
		return
	}
	userIdStr, _ := c.Get("userId")
	userId := userIdStr.(int64)
	//TODO：鉴权
	sendMessageRequest := &im.SendMessageRequest{
		UserId:     util.Int64(userId),
		ConShortId: req.ConShortId,
		ConId:      req.ConId,
		ConType:    req.ConType,
		MsgType:    req.MsgType,
		MsgContent: req.MsgContent,
	}
	sendMessageResponse, err := message.SendMessage(c, sendMessageRequest)
	if err != nil {
		c.JSON(http.StatusOK, sendMessageResponse.BaseResp.StatusCode)
		return
	}
	c.JSON(http.StatusOK, im.StatusCode_Success)
	return
	//检查合法->生成serverMsgId->异步/同步发送消息
	//检查会话(单聊->创建会话(im_conversation_api)，获取会话coreInfo(im_conversation_api，redis+mysql))->检查消息(integration_callback)
	//->处理@消息->发送消息(message_api)->设置最新用户(im_lastuser_api,abase)
	//发送消息(message_api)->参数校验->(同步)redis加锁消息去重->存入数据库(abase/daas)->判断命令消息->写入消息队列->redis解锁
}
