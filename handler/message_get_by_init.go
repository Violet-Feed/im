package handler

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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
	var req *im.GetMessageByInitRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, StateCode_Param_ERROR)
		return
	}
	userIdStr, _ := c.Get("userId")
	userId := userIdStr.(int64)
	userConvIndex := req.GetUserConvIndex()
	if userConvIndex == 0 {
		userConvIndex = math.MaxInt64
	}
	pullUserConvIndexRequest := &im.PullUserConvIndexRequest{
		UserId:        util.Int64(userId),
		UserConvIndex: util.Int64(userConvIndex),
		Limit:         util.Int64(ConvLimit),
	}
	pullUserConvIndexResponse, err := index.PullUserConvIndex(c, pullUserConvIndexRequest)
	if err != nil {
		logrus.Errorf("[GetByInit] PullUserConvIndex err. err = %v", err)
		c.JSON(http.StatusOK, StateCode_Internal_ERROR)
		return
	}
	wg := sync.WaitGroup{}
	for _, conv := range pullUserConvIndexResponse.GetConvShortIds() {
		wg.Add(1)
		go func(ctx context.Context, conv int64) {
			defer wg.Done()
			pullConversationIndexRequest := &im.PullConversationIndexRequest{
				ConvShortId: util.Int64(conv),
				ConvIndex:   util.Int64(math.MaxInt64),
				Limit:       util.Int64(MsgLimit),
			}
			pullConversationIndexResponse, err := index.PullConversationIndex(ctx, pullConversationIndexRequest)
			if err != nil {
				logrus.Errorf("[GetByInit] PullConversationIndex err. err = %v", err)
				return
			}
			getMessagesRequest := &im.GetMessagesRequest{
				ConvShortId: util.Int64(conv),
				MsgIds:      pullConversationIndexResponse.MsgIds,
			}
			getMessagesResponse, err := message.GetMessages(ctx, getMessagesRequest)
			if err != nil {
				logrus.Errorf("[GetByInit] GetMessages err. err = %v", err)
				return
			}
			getMessagesResponse.GetMsgBodies()
		}(c, conv)
	}
	wg.Wait()
	//获取最近会话id(recent_conversation,abase,zset)->拉取会话链(inbox_api,V2)->获取消息内容(message_api)->隐藏撤回消息
	//->获取会话core,setting信息(im_conversation_api)->获取会话ext信息(conversation_ext)->获取群聊是否是成员最近成员信息(im_conversation_api)
	//->获取消息总数(im_counter_manager_rust,mget)->(第一个请求：获取命令链最近index(inbox_api,V2)->获取最近会话version(recent_conversation))
	//->信息整理过滤
}
