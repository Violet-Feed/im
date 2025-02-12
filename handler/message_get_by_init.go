package handler

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"im/handler/conversation"
	"im/handler/index"
	"im/handler/message"
	"im/proto_gen/im"
	"im/util"
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
	userConIndex := req.GetUserConIndex()
	if userConIndex == 0 {
		userConIndex = math.MaxInt64
	}
	pullUserConIndexRequest := &im.PullUserConIndexRequest{
		UserId:       util.Int64(userId),
		UserConIndex: util.Int64(userConIndex),
		Limit:        util.Int64(ConvLimit),
	}
	pullUserConIndexResponse, err := index.PullUserConIndex(c, pullUserConIndexRequest)
	if err != nil {
		logrus.Errorf("[GetByInit] PullUserConIndex err. err = %v", err)
		c.JSON(http.StatusOK, im.StatusCode_Server_Error)
		return
	}
	conShortIds := pullUserConIndexResponse.GetConShortIds()
	wg := sync.WaitGroup{}
	for _, convShortId := range conShortIds {
		wg.Add(1)
		go func(ctx context.Context, convShortId int64) {
			defer wg.Done()
			pullConversationIndexRequest := &im.PullConversationIndexRequest{
				ConShortId: util.Int64(convShortId),
				ConIndex:   util.Int64(math.MaxInt64),
				Limit:      util.Int64(MsgLimit),
			}
			pullConversationIndexResponse, err := index.PullConversationIndex(ctx, pullConversationIndexRequest)
			if err != nil {
				logrus.Errorf("[GetByInit] PullConversationIndex err. err = %v", err)
				return
			}
			getMessageRequest := &im.GetMessagesRequest{
				ConShortId: util.Int64(convShortId),
				MsgIds:     pullConversationIndexResponse.MsgIds,
			}
			getMessageResponse, err := message.GetMessages(ctx, getMessageRequest)
			if err != nil {
				logrus.Errorf("[GetByInit] GetMessage err. err = %v", err)
				return
			}
			getMessageResponse.GetMsgBodies()
		}(c, convShortId)
	}
	wg.Add(1)
	go func(ctx context.Context, userId int64) {
		defer wg.Done()
		pullUserCmdIndexRequest := &im.PullUserCmdIndexRequest{
			UserId:       util.Int64(userId),
			UserCmdIndex: util.Int64(math.MaxInt64),
			Limit:        util.Int64(1),
		}
		pullUserCmdIndexResponse, err := index.PullUserCmdIndex(ctx, pullUserCmdIndexRequest)
		if err != nil {
			logrus.Errorf("[GetByInit] PullUserCmdIndex err. err = %v", err)
			return
		}
		pullUserCmdIndexResponse.GetLastUserCmdIndex()
	}(c, userId)
	wg.Add(1)
	go func(ctx context.Context, convIds []int64) {
		defer wg.Done()
		getConversationBadgeRequest := &im.GetConversationBadgesRequest{
			UserId:      util.Int64(userId),
			ConShortIds: convIds,
		}
		getConversationBadgeResponse, err := conversation.GetConversationBadges(ctx, getConversationBadgeRequest)
		if err != nil {
			logrus.Errorf("[GetByInit] GetConversationBadge err. err = %v", err)
			return
		}
		getConversationBadgeResponse.GetBadgeCounts()
	}(c, conShortIds)
	wg.Add(1)
	go func() {
		defer wg.Done()
		getConversationCoresRequest := &im.GetConversationCoresRequest{
			ConShortIds: conShortIds,
		}
		getConversationCoresResponse, err := conversation.GetConversationCores(c, getConversationCoresRequest)
		if err != nil {
			logrus.Errorf("[GetByInit] GetConversationCores err. err = %v", err)
			return
		}
		getConversationCoresResponse.GetCoreInfos()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		getConversationSettingsRequest := &im.GetConversationSettingsRequest{
			UserId:      util.Int64(userId),
			ConShortIds: conShortIds,
		}
		getConversationSettingsResponse, err := conversation.GetConversationSettings(c, getConversationSettingsRequest)
		if err != nil {
			logrus.Errorf("[GetByInit] GetConversationSettings err. err = %v", err)
			return
		}
		getConversationSettingsResponse.GetSettingInfos()
	}()
	//TODO:获取会话member信息
	wg.Wait()
	//获取最近会话id(recent_conversation,abase,zset)->拉取会话链(inbox_api,V2)->获取消息内容(message_api)->隐藏撤回消息
	//->获取会话core,setting信息(im_conversation_api)->获取会话ext信息(conversation_ext)->获取群聊是否是成员,最近成员信息(im_conversation_api)
	//->获取消息总数(im_counter_manager_rust,mget)->(第一个请求：获取命令链最近index(inbox_api,V2)->获取最近会话version(recent_conversation))
	//->信息整理过滤
}
