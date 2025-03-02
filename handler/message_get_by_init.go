package handler

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"im/biz"
	"im/proto_gen/im"
	"math"
	"net/http"
	"sync"
)

const (
	ConvLimit = 50
	MsgLimit  = 5
)

func GetByInit(c *gin.Context) {
	var req *im.MessageGetByInitRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, im.StatusCode_Param_Error)
		return
	}
	userIdStr, _ := c.Get("userId")
	userId := userIdStr.(int64)
	//拉取用户会话链
	userConIndex := req.GetUserConIndex()
	if userConIndex == 0 {
		userConIndex = math.MaxInt64
	}
	conShortIds, _, _, _, err := biz.PullUserConIndex(c, userId, userConIndex, ConvLimit)
	if err != nil {
		logrus.Errorf("[GetByInit] PullUserConIndex err. err = %v", err)
		c.JSON(http.StatusOK, im.StatusCode_Server_Error)
		return
	}
	//拉取会话链
	wg := sync.WaitGroup{}
	for _, convShortId := range conShortIds {
		wg.Add(1)
		go func(ctx context.Context, convShortId int64) {
			defer wg.Done()
			msgIds, _, err := biz.PullConversationIndex(ctx, convShortId, math.MaxInt64, MsgLimit)
			if err != nil {
				logrus.Errorf("[GetByInit] PullConversationIndex err. err = %v", err)
				return
			}
			_, err = biz.GetMessages(ctx, convShortId, msgIds)
			if err != nil {
				logrus.Errorf("[GetByInit] GetMessage err. err = %v", err)
				return
			}
		}(c, convShortId)
	}
	//获取用户命令链index
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		_, _, err := biz.PullUserCmdIndex(ctx, userId, math.MaxInt64, 1)
		if err != nil {
			logrus.Errorf("[GetByInit] PullUserCmdIndex err. err = %v", err)
			return
		}
	}(c)
	//获取会话badge
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		_, err := biz.GetConversationBadges(ctx, userId, conShortIds)
		if err != nil {
			logrus.Errorf("[GetByInit] GetConversationBadge err. err = %v", err)
			return
		}
	}(c)
	//获取会话core
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		_, err := biz.GetConversationCores(ctx, conShortIds)
		if err != nil {
			logrus.Errorf("[GetByInit] GetConversationCores err. err = %v", err)
			return
		}
	}(c)
	//获取会话setting
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		_, err := biz.GetConversationSettings(ctx, userId, conShortIds)
		if err != nil {
			logrus.Errorf("[GetByInit] GetConversationSettings err. err = %v", err)
			return
		}
	}(c)
	//TODO:获取会话member信息
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		_, err := biz.IsConversationMembers(ctx, conShortIds, userId)
		if err != nil {
			logrus.Errorf("[GetByInit] IsConversationMembers err. err = %v", err)
			return
		}
	}(c)
	//获取最近user_infos
	wg.Add(1)
	go func(ctx context.Context) {
		defer wg.Done()
		_, err := biz.GetConversationMemberInfos(ctx, 0, []int64{0})
		if err != nil {
			logrus.Errorf("[GetByInit] GetConversationUsers err. err = %v", err)
			return
		}
	}(c)
	wg.Wait()
	//获取最近会话id(recent_conversation,abase,zset)->拉取会话链(inbox_api,V2)->获取消息内容(message_api)->隐藏撤回消息
	//->获取会话core,setting信息(im_conversation_api)->获取会话ext信息(conversation_ext)->获取群聊是否是成员,最近成员信息(im_conversation_api)
	//->获取消息总数(im_counter_manager_rust,mget)->(第一个请求：获取命令链最近index(inbox_api,V2)->获取最近会话version(recent_conversation))
	//->信息整理过滤
}
