package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type SendRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func Send(c *gin.Context) {
	var sendRequest SendRequest
	err := c.ShouldBindJSON(&sendRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}
	//检查合法->生成serverMsgId->异步/同步发送消息

	//检查会话(单聊->创建会话(im_conversation_api)，获取会话coreInfo(im_conversation_api，redis+mysql))->检查消息(integration_callback)
	//->处理@消息->发送消息(message_api)->设置最新用户(im_lastuser_api,abase)

	//发送消息(message_api)->参数校验->(同步)redis加锁消息去重->存入数据库(abase/daas)->判断命令消息->写入消息队列->redis解锁

	//消费消息(message_parallel_consumer)->校验过滤->重试消息检查消息是否存在(message_api)->获取strategies->callback->处理ext信息
	//->写会话链(inbox_api,V2)->保存消息->(增加thread未读)->写入用户消息队列->与同步MQ互补->失败发送backup队列

	//消费消息->判断是否为高频用户(本地+redis,写入高频队列batch)->处理重试消息->处理普通消息->更新最近会话(recent,abase,zset)
	//->增加未读(im_counter_manager_rust,abase,xget获取当前已读数和index,redis对msgid加锁,xset)->写入用户链->处理命令消息-
	//>写入命令链(inbox_api,V2)->处理特殊命令->写入用户链->push
}
