package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"im/dal"
	"im/dal/mq"
	"im/proto_gen/common"
	"im/proto_gen/im"
	"im/util"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	ConvLimit = 50
	MsgLimit  = 5
)

func checkMessageSendRequest(req *im.SendMessageRequest) bool {
	if req.GetConId() == "" {
		return false
	}
	if req.GetConType() < 1 || req.GetConType() > 5 {
		return false
	}
	if req.GetMsgType() < 1 || req.GetMsgType() > 5 && req.GetMsgType() != 1000 {
		return false
	}
	if req.GetMsgContent() == "" {
		return false
	}
	return true
}

func SendMessage(ctx context.Context, req *im.SendMessageRequest) (resp *im.SendMessageResponse, err error) {
	resp = &im.SendMessageResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	if !checkMessageSendRequest(req) {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Param_Error, StatusMessage: "参数错误"}
		return resp, errors.New("param error")
	}
	messageId := util.MsgIdGenerator.Generate().Int64()
	//是否群成员
	var isMember int32
	if req.GetConType() == int32(im.ConversationType_One_Chat) {
		isMember = IsSingleMember(ctx, req.GetConId(), req.GetUserId())
	} else if req.GetConType() == int32(im.ConversationType_Group_Chat) {
		status, err := IsGroupsMember(ctx, []int64{req.GetConShortId()}, req.GetUserId())
		if err != nil {
			logrus.Errorf("[SendMessage] IsConversationMembers err. err = %v", err)
			resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
			return resp, err
		}
		isMember = status[0]
	}
	if isMember != 1 {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Not_Found_Error, StatusMessage: "非会话成员"}
		return resp, errors.New("not conversation member")
	}
	if req.GetConType() == int32(im.ConversationType_One_Chat) && req.GetConShortId() == 0 { //创建会话
		parts := strings.Split(req.GetConId(), ":")
		minId, _ := strconv.ParseInt(parts[0], 10, 64)
		maxId, _ := strconv.ParseInt(parts[1], 10, 64)
		createConversationRequest := &im.CreateConversationRequest{
			ConId:   req.GetConId(),
			ConType: req.GetConType(),
			Members: []int64{minId, maxId},
		}
		createConversationResponse, err := CreateConversation(ctx, createConversationRequest)
		if err != nil {
			logrus.Errorf("[SendMessage] CreateConversation err. err = %v", err)
			resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
			return resp, err
		}
		req.ConShortId = createConversationResponse.ConInfo.ConShortId
	}
	//TODO：消息频率控制
	createTime := time.Now().Unix()
	messageBody := &im.MessageBody{
		UserId:      req.GetUserId(),
		ConId:       req.GetConId(),
		ConShortId:  req.GetConShortId(),
		ConType:     req.GetConType(),
		ClientMsgId: req.GetClientMsgId(),
		MsgId:       messageId,
		MsgType:     req.GetMsgType(),
		MsgContent:  req.GetMsgContent(),
		CreateTime:  createTime,
		Extra:       "",
	}
	messageEvent := &im.MessageEvent{
		MsgBody: messageBody,
	}
	err = mq.SendToMq(ctx, "conversation", strconv.FormatInt(req.GetConShortId(), 10), messageEvent)
	if err != nil {
		logrus.Errorf("[SendMessage] SendToMq err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	return resp, nil
}

func GetMessages(ctx context.Context, conShortId int64, msgIds []int64) ([]*im.MessageBody, error) {
	messageBodies := make([]*im.MessageBody, 0)
	if len(msgIds) == 0 {
		return messageBodies, nil
	}
	keys := make([]string, 0)
	for _, id := range msgIds {
		keys = append(keys, fmt.Sprintf("msg:%d:%d", conShortId, id))
	}
	messages, err := dal.KvrocksServer.MGet(ctx, keys)
	if err != nil {
		logrus.Errorf("[GetMessages] kvrocks mget err. err = %v", err)
		return nil, err
	}
	for _, message := range messages {
		var messageBody im.MessageBody
		_ = json.Unmarshal([]byte(message), &messageBody)
		messageBodies = append(messageBodies, &messageBody)
	}
	return messageBodies, nil
}

func StoreMessage(ctx context.Context, msgBody *im.MessageBody) error {
	conShortId := msgBody.GetConShortId()
	messageId := msgBody.GetMsgId()
	key := fmt.Sprintf("msg:%d:%d", conShortId, messageId)
	messageBody, err := json.Marshal(msgBody)
	if err != nil {
		logrus.Errorf("[StoreMessage] marshal messageBody err. err = %v", err)
		return err
	}
	err = dal.KvrocksServer.Set(ctx, key, string(messageBody))
	if err != nil {
		logrus.Errorf("[StoreMessage] kvrocks set err. err = %v", err)
		return err
	}
	return nil
}

func GetMessageByInit(ctx context.Context, req *im.GetMessageByInitRequest) (resp *im.GetMessageByInitResponse, err error) {
	resp = &im.GetMessageByInitResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	userId := req.GetUserId()
	var globalErr error
	conMsgsChan, hasMoreChan, nextUserConIndexChan, userConIndexChan, userCmdIndexChan := make(chan []*im.ConversationMessage), make(chan bool), make(chan int64), make(chan int64), make(chan int64)
	//拉取用户会话链
	go func() {
		userConIndex := req.GetUserConIndex()
		if userConIndex == 0 {
			userConIndex = math.MaxInt64
		}
		conShortIds, userConIndexs, hasMore, err := PullUserConIndex(ctx, userId, userConIndex, ConvLimit)
		if err != nil {
			logrus.Errorf("[GetMessageByInit] PullUserConIndex err. err = %v", err)
			conMsgsChan <- nil
			hasMoreChan <- false
			nextUserConIndexChan <- 0
			userConIndexChan <- 0
			globalErr = err
			return
		}
		hasMoreChan <- hasMore
		if hasMore == true {
			nextUserConIndexChan <- userConIndexs[len(userConIndexs)-1] - 1
		} else {
			nextUserConIndexChan <- 0
		}
		if len(userConIndexs) > 0 {
			userConIndexChan <- userConIndexs[0]
		} else {
			userConIndexChan <- 0
		}
		msgBodiesMapChan, coresMapChan, settingsMapChan := make(chan map[int64][]*im.MessageBody), make(chan map[int64]*im.ConversationCoreInfo), make(chan map[int64]*im.ConversationSettingInfo)
		statusMapChan, badgesMapChan := make(chan map[int64]int32), make(chan map[int64]int64)
		//拉取会话链
		go func() {
			msgBodiesChan := make(chan []*im.MessageBody, len(conShortIds))
			for _, convShortId := range conShortIds {
				go func(convShortId int64) {
					msgIds, conIndexs, err := PullConversationIndex(ctx, convShortId, math.MaxInt64, MsgLimit)
					if err != nil {
						logrus.Errorf("[GetMessageByInit] PullConversationIndex err. err = %v", err)
						msgBodiesChan <- nil
						return
					}
					//TODO:过滤撤回消息
					msgBodies, err := GetMessages(ctx, convShortId, msgIds)
					if err != nil {
						logrus.Errorf("[GetMessageByInit] GetMessage err. err = %v", err)
						msgBodiesChan <- nil
						return
					}
					for i, msgBody := range msgBodies {
						msgBody.ConIndex = conIndexs[i]
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
			cores, err := GetConversationCores(ctx, conShortIds)
			if err != nil {
				logrus.Errorf("[GetMessageByInit] GetConversationCores err. err = %v", err)
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
						statusMap[conShortId] = IsSingleMember(ctx, core.GetConId(), userId)
					} else if core.GetConType() == int32(im.ConversationType_Group_Chat) {
						groupIds = append(groupIds, conShortId)
					}
				}
				status, err := IsGroupsMember(ctx, groupIds, userId)
				if err != nil {
					logrus.Errorf("[GetMessageByInit] IsConversationMembers err. err = %v", err)
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
				_, err := GetConversationMemberInfos(ctx, 0, []int64{0})
				if err != nil {
					logrus.Errorf("[GetMessageByInit] GetConversationUsers err. err = %v", err)
					return
				}
			}()
		}()
		//获取会话setting
		go func() {
			settings, err := GetConversationSettings(ctx, userId, conShortIds)
			if err != nil {
				logrus.Errorf("[GetMessageByInit] GetConversationSettings err. err = %v", err)
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
			badges, err := GetConversationBadges(ctx, userId, conShortIds)
			if err != nil {
				logrus.Errorf("[GetMessageByInit] GetConversationBadge err. err = %v", err)
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
				ConShortId:     core.GetConShortId(),
				ConId:          core.GetConId(),
				ConType:        core.GetConType(),
				UserConIndex:   userConIndexs[i],
				BadgeCount:     badgesMap[conShortId],
				IsMember:       statusMap[conShortId] == 1,
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
		_, userCmdIndex, err := PullUserCmdIndex(ctx, userId, math.MaxInt64, 1)
		if err != nil {
			logrus.Errorf("[GetMessageByInit] PullUserCmdIndex err. err = %v", err)
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
		logrus.Errorf("[GetMessageByInit] happend err.")
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: globalErr.Error()}
		return resp, globalErr
	}
	resp.Cons = conMsgs
	resp.HasMore = hasMore
	resp.NextUserConIndex = nextUserConIndex
	resp.UserConIndex = userConIndex
	resp.UserCmdIndex = userCmdIndex
	return resp, nil
	//获取最近会话id(recent_conversation,abase,zset)->拉取会话链(inbox_api,V2)->获取消息内容(message_api)->隐藏撤回消息
	//->获取会话core,setting信息(im_conversation_api)->获取会话ext信息(conversation_ext)->获取群聊是否是成员,最近成员信息(im_conversation_api)
	//->获取消息总数(im_counter_manager_rust,mget)->(第一个请求：获取命令链最近index(inbox_api,V2)->获取最近会话version(recent_conversation))
	//->信息整理过滤
}

func GetMessageByConversation(ctx context.Context, req *im.GetMessageByConversationRequest) (resp *im.GetMessageByConversationResponse, err error) {
	resp = &im.GetMessageByConversationResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	userId := req.GetUserId()
	cores, err := GetConversationCores(ctx, []int64{req.GetConShortId()})
	if err != nil {
		logrus.Errorf("[GetMessageByConversation] GetConversationCores err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	if len(cores) == 0 || cores[0].GetStatus() != 0 {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Not_Found_Error, StatusMessage: "会话不存在"}
		return resp, errors.New("conversation not found")
	}
	//是否群成员
	var isMember int32
	if cores[0].GetConType() == int32(im.ConversationType_One_Chat) {
		isMember = IsSingleMember(ctx, cores[0].GetConId(), userId)
	} else if cores[0].GetConType() == int32(im.ConversationType_Group_Chat) {
		status, err := IsGroupsMember(ctx, []int64{req.GetConShortId()}, userId)
		if err != nil {
			logrus.Errorf("[GetMessageByConversation] IsGroupsMember err. err = %v", err)
			resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
			return resp, err
		}
		isMember = status[0]
	}
	if isMember != 1 {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Not_Found_Error, StatusMessage: "非会话成员"}
		return resp, errors.New("not conversation member")
	}
	//拉取会话链
	msgIds, conIndexs, err := PullConversationIndex(ctx, req.GetConShortId(), req.GetConIndex(), req.GetLimit())
	if err != nil {
		logrus.Errorf("[GetMessageByConversation] PullConversationIndex err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	msgBodies, err := GetMessages(ctx, req.GetConShortId(), msgIds)
	if err != nil {
		logrus.Errorf("[GetMessageByConversation] GetMessages err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error, StatusMessage: err.Error()}
		return resp, err
	}
	for i, msgBody := range msgBodies {
		msgBody.ConIndex = conIndexs[i]
	}
	resp.MsgBodies = msgBodies
	return resp, nil
	//获取core信息(mysql)->获取成员数量(redis+mysql)->判断是否为成员(redis,mysql)->拉取会话链(loadmore)->隐藏撤回消息
}

func GetMessageByUser(ctx context.Context) {
	//是否有stranger消息(conversation_rust)->拉取命令链(inbox_api,V2)->获取最近读会话及已读index(recent,im_conversation_api)->获取咨询、通知会话(recent_conversation)
}
