package handler

import (
	"github.com/gin-gonic/gin"
)

func GetByUser(c *gin.Context) {
	//是否有stranger消息(conversation_rust)->拉取命令链(inbox_api,V2)->获取最近读会话及已读index(recent,im_conversation_api)->获取咨询、通知会话(recent_conversation)
}
