package handler

import "github.com/gin-gonic/gin"

func GetByConv(c *gin.Context) {
	//获取core信息(mysql)->获取成员数量(redis+mysql)->判断是否为成员(redis,mysql)->拉取会话链(loadmore)->隐藏撤回消息
}
