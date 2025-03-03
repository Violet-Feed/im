package handler

import (
	"github.com/gin-gonic/gin"
	"im/biz"
	"im/proto_gen/im"
	"im/util"
	"net/http"
)

func GetByConversation(c *gin.Context) {
	resp := &im.MessageGetByConversationResponse{}
	var req *im.MessageGetByConversationRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, im.StatusCode_Param_Error)
		return
	}
	userIdStr, _ := c.Get("userId")
	userId := userIdStr.(int64)

	cores,err:=biz.GetConversationCores(c,[]int64{req.GetConShortId()})
	if err!=nil{
		c.JSON(http.StatusOK,im.StatusCode_Server_Error)
		return
	}
	if len(cores)==0||cores[0].GetStatus()!=0{
		c.JSON(http.StatusOK,im.StatusCode_Not_Found_Error)
		return
	}
	//是否群成员
	var isMember int32
	if cores[0].GetConType() == int32(im.ConversationType_One_Chat) {
		isMember = biz.IsSingleMember(c, cores[0].GetConId(), userId)
	} else if cores[0].GetConType() == int32(im.ConversationType_Group_Chat) {
		status, err := biz.IsGroupsMember(c, []int64{req.GetConShortId()}, userId)
		if err != nil {
			c.JSON(http.StatusOK,im.StatusCode_Server_Error)
			return
		}
		isMember = status[0]
	}
	if isMember != 1 {
		c.JSON(http.StatusOK,im.StatusCode_Not_Found_Error)
		return
	}
	//拉取会话链
	msgIds, conIndexs, err := biz.PullConversationIndex(c, req.GetConShortId(), req.GetConIndex(), req.GetLimit())
	if err != nil {
		c.JSON(http.StatusOK,im.StatusCode_Server_Error)
		return
	}
	msgBodies, err := biz.GetMessages(c,  req.GetConShortId(), msgIds)
	if err != nil {
		c.JSON(http.StatusOK,im.StatusCode_Server_Error)
		return
	}
	for i, msgBody := range msgBodies {
		msgBody.ConIndex = util.Int64(conIndexs[i])
	}
	resp.MsgBodies = msgBodies
	c.JSON(http.StatusOK,HttpResponse{
		Code:    im.StatusCode_Success,
		Message: "success",
		Data:    resp,
	})
	return
	//获取core信息(mysql)->获取成员数量(redis+mysql)->判断是否为成员(redis,mysql)->拉取会话链(loadmore)->隐藏撤回消息
}
