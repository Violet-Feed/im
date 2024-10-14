package handler

import (
	"github.com/gin-gonic/gin"
	"im/dal/mq"
	"im/handler/conversation"
	"im/proto_gen/im"
	"im/util"
	"net/http"
	"time"
)

func checkMessageSendRequest(c *gin.Context, req *im.SendMessageRequest) bool {
	//TODO：参数校验
	return true
}

func Send(c *gin.Context) {
	var sendMessageRequest *im.SendMessageRequest
	err := c.ShouldBindJSON(&sendMessageRequest)
	if err != nil || !checkMessageSendRequest(c, sendMessageRequest) {
		c.JSON(http.StatusOK, StateCode_Param_ERROR)
		return
	}
	id, _ := c.Get("userId")
	userId := id.(int64)

	//TODO：鉴权
	messageId := util.MsgIdGenerator.Generate().Int64()
	if sendMessageRequest.GetConvType() == int32(im.ConversationType_ConversationType_One_Chat) {
		if sendMessageRequest.GetConvShortId() == 0 { //创建会话
			createConversationRequest := &im.CreateConversationRequest{
				ConvId:   sendMessageRequest.ConvId,
				ConvType: sendMessageRequest.ConvType,
				OwnerId:  util.Int64(userId),
			}
			createConversationResponse := &im.CreateConversationResponse{}
			err := conversation.CreateConversation(c, createConversationRequest, createConversationResponse)
			if err != nil {
				c.JSON(http.StatusOK, StateCode_Internal_ERROR)
				return
			}
			sendMessageRequest.ConvShortId = createConversationResponse.ConvShortId
		}
	}
	createTime := time.Now().UnixMilli()
	//TODO：消息频率控制
	messageEvent := &im.MessageEvent{
		UserId:      util.Int64(userId),
		ConvId:      sendMessageRequest.ConvId,
		ConvShortId: sendMessageRequest.ConvShortId,
		ConvType:    sendMessageRequest.ConvType,
		MsgId:       util.Int64(messageId),
		MsgType:     sendMessageRequest.MsgType,
		MsgContent:  sendMessageRequest.MsgContent,
		CreateTime:  util.Int64(createTime),
	}
	err = mq.SendMq(c, "conversation", "", messageEvent)
	if err != nil {
		c.JSON(http.StatusOK, StateCode_Internal_ERROR)
		return
	}
	c.JSON(http.StatusOK, StatusCode_Success)
	return
	//检查合法->生成serverMsgId->异步/同步发送消息
	//检查会话(单聊->创建会话(im_conversation_api)，获取会话coreInfo(im_conversation_api，redis+mysql))->检查消息(integration_callback)
	//->处理@消息->发送消息(message_api)->设置最新用户(im_lastuser_api,abase)
	//发送消息(message_api)->参数校验->(同步)redis加锁消息去重->存入数据库(abase/daas)->判断命令消息->写入消息队列->redis解锁
}
