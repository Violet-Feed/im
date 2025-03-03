package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"im/biz"
	"im/proto_gen/im"
	"im/util"
	"math"
	"net/http"
)

const (
	ConvLimit = 50
	MsgLimit  = 5
)

func GetByInit(c *gin.Context) {
	resp := &im.MessageGetByInitResponse{}
	var req *im.MessageGetByInitRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(http.StatusOK, im.StatusCode_Param_Error)
		return
	}
	userIdStr, _ := c.Get("userId")
	userId := userIdStr.(int64)

	var globalErr error
	conMsgsChan, hasMoreChan, nextUserConIndexChan, userConIndexChan, userCmdIndexChan := make(chan []*im.ConversationMessage), make(chan bool), make(chan int64), make(chan int64), make(chan int64)
	//拉取用户会话链
	go func() {
		userConIndex := req.GetUserConIndex()
		if userConIndex == 0 {
			userConIndex = math.MaxInt64
		}
		conShortIds, userConIndexs, hasMore, err := biz.PullUserConIndex(c, userId, userConIndex, ConvLimit)
		if err != nil {
			logrus.Errorf("[GetByInit] PullUserConIndex err. err = %v", err)
			conMsgsChan <- nil
			hasMoreChan <- false
			nextUserConIndexChan <- 0
			userConIndexChan <- 0
			globalErr = err
			return
		}
		hasMoreChan <- hasMore
		if userConIndexs != nil {
			nextUserConIndexChan <- userConIndexs[len(userConIndexs)-1] - 1
			userConIndexChan <- userConIndexs[0]
		} else {
			nextUserConIndexChan <- 0
			userConIndexChan <- 0
		}
		msgBodiesMapChan, coresMapChan, settingsMapChan := make(chan map[int64][]*im.MessageBody), make(chan map[int64]*im.ConversationCoreInfo), make(chan map[int64]*im.ConversationSettingInfo)
		statusMapChan, badgesMapChan := make(chan map[int64]int32), make(chan map[int64]int64)
		//拉取会话链
		go func() {
			msgBodiesChan := make(chan []*im.MessageBody, len(conShortIds))
			for _, convShortId := range conShortIds {
				go func(convShortId int64) {
					msgIds, conIndexs, err := biz.PullConversationIndex(c, convShortId, math.MaxInt64, MsgLimit)
					if err != nil {
						logrus.Errorf("[GetByInit] PullConversationIndex err. err = %v", err)
						msgBodiesChan <- nil
						return
					}
					//TODO:过滤撤回消息
					msgBodies, err := biz.GetMessages(c, convShortId, msgIds)
					if err != nil {
						logrus.Errorf("[GetByInit] GetMessage err. err = %v", err)
						msgBodiesChan <- nil
						return
					}
					for i, msgBody := range msgBodies {
						msgBody.ConIndex = util.Int64(conIndexs[i])
					}
					msgBodiesChan <- msgBodies
				}(convShortId)
			}
			msgBodiesMap := make(map[int64][]*im.MessageBody)
			for i := 0; i < len(conShortIds); i++ {
				msgBodies := <-msgBodiesChan
				if msgBodies != nil {
					msgBodiesMap[msgBodies[0].GetConShortId()] = msgBodies
				}
			}
			msgBodiesMapChan <- msgBodiesMap
		}()
		//获取会话core
		go func() {
			cores, err := biz.GetConversationCores(c, conShortIds)
			if err != nil {
				logrus.Errorf("[GetByInit] GetConversationCores err. err = %v", err)
				coresMapChan <- nil
				statusMapChan <- nil
				globalErr = err
				return
			}
			coresMap := make(map[int64]*im.ConversationCoreInfo)
			for _, core := range cores {
				coresMap[core.GetConShortId()] = core
			}
			coresMapChan <- coresMap
			//获取是否是会话member
			go func() {
				statusMap := make(map[int64]int32)
				groupIds := make([]int64, 0)
				for _, conShortId := range conShortIds {
					core := coresMap[conShortId]
					if core.GetConType() == int32(im.ConversationType_One_Chat) {
						statusMap[conShortId] = biz.IsSingleMember(c, core.GetConId(), userId)
					} else if core.GetConType() == int32(im.ConversationType_Group_Chat) {
						groupIds = append(groupIds, conShortId)
					}
				}
				status, err := biz.IsGroupsMember(c, groupIds, userId)
				if err != nil {
					logrus.Errorf("[GetByInit] IsConversationMembers err. err = %v", err)
					statusMapChan <- statusMap
					return
				}
				for k, v := range status {
					statusMap[k] = v
				}
				statusMapChan <- statusMap
			}()
			//TODO:获取最近member，id昵称权限block
			go func() {
				_, err := biz.GetConversationMemberInfos(c, 0, []int64{0})
				if err != nil {
					logrus.Errorf("[GetByInit] GetConversationUsers err. err = %v", err)
					return
				}
			}()
		}()
		//获取会话setting
		go func() {
			settings, err := biz.GetConversationSettings(c, userId, conShortIds)
			if err != nil {
				logrus.Errorf("[GetByInit] GetConversationSettings err. err = %v", err)
				settingsMapChan <- nil
				globalErr = err
				return
			}
			settingsMap := make(map[int64]*im.ConversationSettingInfo)
			for _, setting := range settings {
				settingsMap[setting.GetConShortId()] = setting
			}
			settingsMapChan <- settingsMap
		}()
		//获取会话badge
		go func() {
			badges, err := biz.GetConversationBadges(c, userId, conShortIds)
			if err != nil {
				logrus.Errorf("[GetByInit] GetConversationBadge err. err = %v", err)
				badgesMapChan <- nil
				return
			}
			badgesMap := make(map[int64]int64)
			for i, conShortId := range conShortIds {
				badgesMap[conShortId] = badges[i]
			}
			badgesMapChan <- badgesMap
		}()
		msgBodiesMap := <-msgBodiesMapChan
		coresMap := <-coresMapChan
		settingsMap := <-settingsMapChan
		statusMap := <-statusMapChan
		badgesMap := <-badgesMapChan
		conMsgs := make([]*im.ConversationMessage, 0)
		//TODO:通过minIndex过滤消息,过滤非成员，无core、setting
		for i, conShortId := range conShortIds {
			core := coresMap[conShortId]
			conInfo := &im.ConversationInfo{
				ConShortId:     core.ConShortId,
				ConId:          core.ConId,
				ConType:        core.ConType,
				UserConIndex:   util.Int64(userConIndexs[i]),
				BadgeCount:     util.Int64(badgesMap[conShortId]),
				IsMember:       util.Bool(statusMap[conShortId] == 1),
				ConCoreInfo:    core,
				ConSettingInfo: settingsMap[conShortId],
			}
			conMsg := &im.ConversationMessage{
				ConInfo:   conInfo,
				MsgBodies: msgBodiesMap[conShortId],
			}
			conMsgs = append(conMsgs, conMsg)
		}
		conMsgsChan <- conMsgs
	}()
	//获取用户命令链index
	go func() {
		if req.GetUserConIndex() != 0 {
			userCmdIndexChan <- 0
			return
		}
		_, userCmdIndex, err := biz.PullUserCmdIndex(c, userId, math.MaxInt64, 1)
		if err != nil {
			logrus.Errorf("[GetByInit] PullUserCmdIndex err. err = %v", err)
			userCmdIndexChan <- 0
			globalErr = err
			return
		}
		userCmdIndexChan <- userCmdIndex
	}()
	hasMore := <-hasMoreChan
	nextUserConIndex := <-nextUserConIndexChan
	userConIndex := <-userConIndexChan
	userCmdIndex := <-userCmdIndexChan
	conMsgs := <-conMsgsChan
	if globalErr != nil {
		logrus.Errorf("[GetByInit] happend err.")
		c.JSON(http.StatusOK, HttpResponse{
			Code:    im.StatusCode_Server_Error,
			Message: "server err",
			Data:    resp,
		})
		return
	}
	resp.Cons = conMsgs
	resp.HasMore = util.Bool(hasMore)
	resp.NextUserConIndex = util.Int64(nextUserConIndex)
	resp.UserConIndex = util.Int64(userConIndex)
	resp.UserCmdIndex = util.Int64(userCmdIndex)
	c.JSON(http.StatusOK, HttpResponse{
		Code:    im.StatusCode_Success,
		Message: "success",
		Data:    resp,
	})
	return
	//获取最近会话id(recent_conversation,abase,zset)->拉取会话链(inbox_api,V2)->获取消息内容(message_api)->隐藏撤回消息
	//->获取会话core,setting信息(im_conversation_api)->获取会话ext信息(conversation_ext)->获取群聊是否是成员,最近成员信息(im_conversation_api)
	//->获取消息总数(im_counter_manager_rust,mget)->(第一个请求：获取命令链最近index(inbox_api,V2)->获取最近会话version(recent_conversation))
	//->信息整理过滤
}
