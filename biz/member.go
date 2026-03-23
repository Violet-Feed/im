package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"im/biz/model"
	"im/dal"
	"im/proto_gen/common"
	"im/proto_gen/im"
	"math"
	"strconv"
	"strings"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const ConversationLimit = 100

func AddConversationMembers(ctx context.Context, req *im.AddConversationMembersRequest) (resp *im.AddConversationMembersResponse, err error) {
	//目前采用先操作再发消息的方案，可能会出现新成员会看到入群消息之前消息的问题
	resp = &im.AddConversationMembersResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	conShortId := req.GetConShortId()
	//判断群是否存在
	cores, err := GetConversationCores(ctx, []int64{conShortId}, false)
	if err != nil {
		logrus.Errorf("[AddConversationMembers] GetConversationCores err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, err
	}
	if len(cores) == 0 || cores[0] == nil || cores[0].GetStatus() != 0 || cores[0].GetConType() != int32(im.ConversationType_Group_Chat) {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Not_Found_Error, StatusMessage: "会话不存在"}
		return resp, nil
	}
	//获取成员数量
	locked := dal.RedisServer.Lock(ctx, fmt.Sprintf("user_count:%d", conShortId))
	if !locked {
		logrus.Errorf("[AddConversationMembers] Lock err.")
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_OverFrequency_Error, StatusMessage: "访问频繁"}
		return resp, err
	}
	defer dal.RedisServer.Unlock(ctx, fmt.Sprintf("user_count:%d", conShortId))
	count, err := model.GetUserCount(ctx, conShortId)
	if err != nil {
		logrus.Errorf("[AddConversationMembers] GetUserCount err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, err
	}
	if count+len(req.GetMembers()) > ConversationLimit {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_OverLimit_Error, StatusMessage: "成员数量达到上限"}
		return resp, nil
	}
	//创建userInfo
	var userModels []*model.ConversationUserInfo
	for _, member := range req.GetMembers() {
		userModel := model.PackUserModel(member, req)
		if count == 0 && member == req.GetOperator() {
			userModel.Privilege = 1
		}
		userModels = append(userModels, userModel)
	}
	err = model.InsertUserInfos(ctx, conShortId, userModels)
	if err != nil {
		logrus.Errorf("[AddConversationMembers] InsertUserInfos err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, err
	}
	//获取设置index
	_, conIndex, err := PullConversationIndex(ctx, conShortId, math.MaxInt64, 1)
	if err != nil {
		logrus.Errorf("[AddConversationMembers] PullConversationIndex err. err = %v", err)
	} else {
		minIndex := int64(0)
		if len(conIndex) > 0 {
			minIndex = conIndex[0]
		}
		err = model.SetReadIndexStart(ctx, conShortId, req.GetMembers(), minIndex)
		if err != nil {
			logrus.Errorf("[AddConversationMembers] SetReadIndexStart err. err = %v", err)
		}
	}
	//发送进群命令消息
	conMessage := map[string]interface{}{
		"type":     im.ConMessageType_Add_Member,
		"operator": req.GetOperator(),
		"content":  req.GetMembers(),
	}
	conMessageByte, _ := json.Marshal(conMessage)
	sendMessageRequest := &im.SendMessageRequest{
		SenderId:    0,
		SenderType:  int32(im.SenderType_Conv),
		ConShortId:  req.GetConShortId(),
		ConId:       req.GetConId(),
		ConType:     int32(im.ConversationType_Group_Chat),
		MsgType:     int32(im.MessageType_Conversation),
		MsgContent:  string(conMessageByte),
		ClientMsgId: 0,
	}
	logrus.Infof("[AddConversationMembers] sendMessageRequest = %v", sendMessageRequest)
	_, err = SendMessage(ctx, sendMessageRequest)
	if err != nil {
		logrus.Errorf("[AddConversationMembers] SendMessage err. err = %v", err)
	}
	//个人认为流程：redis加锁，获取成员数量判断是否超限，存入数据库，发送普通消息，解锁，入单链，对于新成员设置已读起点终点minIndex，入用户链，对于新成员获取badge设置readBadge
	//再次入群获取保存badgeCount->判断成员数量limit->保存userModels->再次判断limit，删除成员
	//拉会话链，获取index起点,设置已读起点->发送进群命令消息->发送挡板消息
	return resp, nil
}

func GetConversationMembers(ctx context.Context, req *im.GetConversationMembersRequest) (*im.GetConversationMembersResponse, error) {
	resp := &im.GetConversationMembersResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	userIds, err := GetConversationMemberIds(ctx, req.GetConShortId())
	if err != nil {
		logrus.Errorf("[GetConversationMembers] GetConversationMemberIds err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, nil
	}
	userInfos, err := GetConversationMemberInfos(ctx, req.GetConShortId(), userIds)
	if err != nil {
		logrus.Errorf("[GetConversationMembers] GetConversationMemberInfos err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, nil
	}
	resp.Members = userInfos
	return resp, nil
}

func GetConversationMemberIds(ctx context.Context, conShortId int64) ([]int64, error) {
	userIds, err := model.GetUserIdList(ctx, conShortId)
	if err != nil {
		logrus.Errorf("[GetConversationMembers] GetUserIdList err. err = %v", err)
		return nil, err
	}
	return userIds, nil
}

func GetConversationMemberInfos(ctx context.Context, conShortId int64, userIds []int64) ([]*im.ConversationUserInfo, error) {
	var userInfos []*im.ConversationUserInfo
	if len(userIds) == 0 {
		return userInfos, nil
	}
	userMap, err := model.GetUserInfos(ctx, conShortId, userIds, true)
	if err != nil {
		logrus.Errorf("[GetConversationMemberInfos] GetUserInfos err. err = %v", err)
		return nil, err
	}
	for _, id := range userIds {
		userInfos = append(userInfos, model.PackUserInfo(userMap[id]))
	}
	return userInfos, nil
}

func IsGroupsMember(ctx context.Context, conShortIds []int64, userId int64) (map[int64]int32, error) {
	wg := sync.WaitGroup{}
	statusChan := make([]chan int32, len(conShortIds))
	for i := range statusChan {
		statusChan[i] = make(chan int32, 1)
	}
	for i, conShortId := range conShortIds {
		wg.Add(1)
		go func(i int, conShortId int64) {
			defer wg.Done()
			_, err := dal.RedisServer.ZScore(ctx, "member:"+strconv.FormatInt(conShortId, 10), strconv.FormatInt(userId, 10))
			if err == nil {
				statusChan[i] <- 1
				return
			} else if errors.Is(err, redis.Nil) {
				statusChan[i] <- 0
				return
			} else {
				userInfo, err := model.GetUserInfos(ctx, conShortId, []int64{userId}, false)
				if err != nil {
					logrus.Errorf("[IsMembers] GetUserInfos err. err = %v", err)
					statusChan[i] <- -1
					return
				}
				if len(userInfo) > 0 {
					statusChan[i] <- 1
				} else {
					statusChan[i] <- 0
				}
			}
		}(i, conShortId)
	}
	wg.Wait()
	status := make(map[int64]int32)
	for i, conShortId := range conShortIds {
		status[conShortId] = <-statusChan[i]
	}
	return status, nil
	//获取core？
	//并发redis zscore;err mysql
}

func IsGroupAI(ctx context.Context, conShortId int64, agentId int64) (int32, error) {
	agentInfo, err := model.GetAgentInfosByIds(ctx, conShortId, []int64{agentId})
	if err != nil {
		logrus.Errorf("[IsGroupAI] GetAgentInfosByIds err. err = %v", err)
		return 0, err
	}
	if len(agentInfo) > 0 {
		return 1, nil
	}
	return 0, nil
}

func IsSingleMember(ctx context.Context, conId string, userId int64) int32 {
	parts := strings.Split(conId, ":")
	minId, _ := strconv.ParseInt(parts[0], 10, 64)
	maxId, _ := strconv.ParseInt(parts[1], 10, 64)
	if userId == minId || userId == maxId {
		return 1
	}
	return 0
}

func IsAIMember(ctx context.Context, conId string, senderType int32, senderId int64) int32 {
	parts := strings.Split(conId, ":")
	realId, _ := strconv.ParseInt(parts[1], 10, 64)
	if senderType == int32(im.SenderType_AI) {
		realId, _ = strconv.ParseInt(parts[2], 10, 64)
	}
	if senderId == realId {
		return 1
	}
	return 0
}

func GetMembersReadIndex(ctx context.Context, req *im.GetMembersReadIndexRequest) (resp *im.GetMembersReadIndexResponse, err error) {
	resp = &im.GetMembersReadIndexResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	userIds, err := GetConversationMemberIds(ctx, req.GetConShortId())
	if err != nil {
		logrus.Errorf("[GetMembersReadIndex] GetConversationMemberIds err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, nil
	}
	readIndexes, err := model.GetMemberReadIndexEnd(ctx, req.GetConShortId(), userIds)
	if err != nil {
		logrus.Errorf("[GetMembersReadIndex] GetMemberReadIndexEnd err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, nil
	}
	resp.ReadIndex = readIndexes
	return resp, nil
}
