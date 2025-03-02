package handler

import "github.com/gin-gonic/gin"

func MarkRead(c *gin.Context) {
	//设置总、各会话未读数(im_counter_manager_rust)->设置已读index和count(im_conversation_api,redis hash,abase xset)->
	//发送命令消息->更新最近会话(recent)->lastindex,保存,写命令链,push->发送已读消息(im_lastuser_api)
}
