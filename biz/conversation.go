package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"im/biz/model"
	"im/dal"
	"im/proto_gen/common"
	"im/proto_gen/im"
	"im/util"
	"strconv"
	"time"
)

func CreateConversation(ctx context.Context, req *im.CreateConversationRequest) (resp *im.CreateConversationResponse, err error) {
	resp = &im.CreateConversationResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	conShortId := util.ConIdGenerator.Generate().Int64()
	if req.GetConType() == int32(im.ConversationType_One_Chat) || req.GetConType() == int32(im.ConversationType_AI_Chat) {
		//对conId幂等，目前直接查询，考虑创建Identity
		core, err := model.GetCoreInfoByConId(ctx, req.GetConId())
		if err != nil {
			logrus.Errorf("[CreateConversation] GetCoreInfoByConId err. err = %v", err)
			resp.BaseResp.StatusCode = common.StatusCode_Server_Error
			return resp, err
		}
		if core != nil {
			resp.ConCoreInfo = model.PackCoreInfo(core)
			return resp, nil
		}
	}
	if req.GetConType() == int32(im.ConversationType_Group_Chat) {
		req.ConId = strconv.FormatInt(conShortId, 10)
	}
	coreModel := model.PackCoreModel(conShortId, req)
	//创建core
	err = model.InsertCoreInfo(ctx, coreModel)
	if err != nil {
		logrus.Errorf("[CreateConversation] InsertCoreInfo err. err = %v", err)
		resp.BaseResp.StatusCode = common.StatusCode_Server_Error
		return resp, err
	}
	if req.GetConType() == int32(im.ConversationType_One_Chat) || req.GetConType() == int32(im.ConversationType_AI_Chat) {
		//创建setting
		for _, member := range req.GetMembers() {
			if member == int64(common.SpecialUser_AI) {
				continue
			}
			settingModel := model.PackSettingModel(member, conShortId, req)
			err := model.InsertSettingInfo(ctx, settingModel)
			if err != nil {
				logrus.Errorf("[CreateConversation] InsertSettingInfo err. err = %v", err)
				//失败不return，如果查询时没有setting再创建
			}
		}
	}
	if req.GetConType() == int32(im.ConversationType_Group_Chat) {
		//添加成员
		_, err := AddConversationMembers(ctx, &im.AddConversationMembersRequest{
			ConShortId: conShortId,
			ConId:      req.GetConId(),
			Members:    req.GetMembers(),
			Operator:   req.GetOwnerId(),
		})
		if err != nil {
			logrus.Errorf("[CreateConversation] AddConversationMembers err. err = %v", err)
			resp.BaseResp.StatusCode = common.StatusCode_Server_Error
			return resp, err
		}
		//暂时不发命令消息
	}
	coreInfo := model.PackCoreInfo(coreModel)
	coreInfo.MemberCount = int32(len(req.GetMembers()))
	resp.ConCoreInfo = coreInfo
	return resp, nil
	//(幂等创建Identity->写redis->)创建/更新core->写redis->发送命令消息，单聊更新setting，群聊添加成员、审核开关
}

func GetConversationInfo(ctx context.Context, req *im.GetConversationInfoRequest) (resp *im.GetConversationInfoResponse, err error) {
	resp = &im.GetConversationInfoResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	core, err := GetConversationCores(ctx, []int64{req.GetConShortId()}, true)
	if len(core) == 0 || core[0] == nil || core[0].GetStatus() != 0 {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Not_Found_Error, StatusMessage: "会话不存在"}
		return resp, nil
	}
	if err != nil {
		logrus.Errorf("[GetConversationInfo] GetConversationCores err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, nil
	}
	setting, err := GetConversationSettings(ctx, req.GetUserId(), []int64{req.GetConShortId()})
	if err != nil {
		logrus.Errorf("[GetConversationInfo] GetConversationSettings err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, nil
	}
	userIds, err := GetConversationMemberIds(ctx, req.GetConShortId())
	if err != nil {
		logrus.Errorf("[GetConversationInfo] GetConversationMemberIds err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, nil
	}
	userInfos, err := GetConversationMemberInfos(ctx, req.GetConShortId(), userIds)
	if err != nil {
		logrus.Errorf("[GetConversationInfo] GetConversationMemberInfos err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, nil
	}
	conInfo := &im.ConversationInfo{
		ConShortId:     req.GetConShortId(),
		ConId:          core[0].GetConId(),
		ConType:        core[0].GetConType(),
		UserConIndex:   0,
		BadgeCount:     0,
		IsMember:       true,
		Members:        userInfos,
		ConCoreInfo:    core[0],
		ConSettingInfo: setting[0],
	}
	resp.ConInfo = conInfo
	return resp, nil
}

func GetConversationCores(ctx context.Context, conShortIds []int64, needUserCount bool) ([]*im.ConversationCoreInfo, error) {
	coreInfos := make([]*im.ConversationCoreInfo, 0)
	if len(conShortIds) == 0 {
		return coreInfos, nil
	}
	coresMap, err := model.GetCoreInfos(ctx, conShortIds)
	if err != nil {
		logrus.Errorf("[GetConversationCores] GetCoreInfos err. err = %v", err)
		return nil, err
	}
	for _, id := range conShortIds {
		coreInfos = append(coreInfos, model.PackCoreInfo(coresMap[id]))
	}
	if needUserCount {
		countChan := make([]chan int, len(conShortIds))
		for i := range countChan {
			countChan[i] = make(chan int)
		}
		for i, conShortId := range conShortIds {
			go func(i int, conShortId int64) {
				count, err := model.GetUserCount(ctx, conShortId)
				if err != nil {
					logrus.Errorf("[GetConversationCores] GetUserCount err. err = %v", err)
					countChan[i] <- 0
				}
				countChan[i] <- count
			}(i, conShortId)
		}
		for i := 0; i < len(conShortIds); i++ {
			if coreInfos[i] == nil {
				<-countChan[i]
			}
			coreInfos[i].MemberCount = int32(<-countChan[i])
		}
	}
	return coreInfos, nil
	//redis mget key:convId,mysql,redis一天
	//并发获取成员数量：useCache:本地缓存key:convId,get,redis get;nil or noUse:redis key:convId,zcard;nil mysql 设置string一分钟,zset永久
}

func GetConversationSettings(ctx context.Context, userId int64, conShortIds []int64) ([]*im.ConversationSettingInfo, error) {
	settingInfos := make([]*im.ConversationSettingInfo, 0)
	if len(conShortIds) == 0 {
		return settingInfos, nil
	}
	settingModel, err := model.GetSettingInfo(ctx, userId, conShortIds)
	if err != nil {
		logrus.Errorf("[GetConversationSettings] GetSettingInfo err. err = %v", err)
		return nil, err
	}
	var missIds []int64
	for _, id := range conShortIds {
		if settingModel[id] == nil {
			missIds = append(missIds, id)
		}
	}
	if len(missIds) != 0 {
		fillModel, err := fillSettingModel(ctx, userId, missIds)
		if err != nil {
			logrus.Errorf("[GetConversationSettings] FixSettingModel err. err = %v", err)
		} else {
			for k, v := range fillModel {
				settingModel[k] = v
			}
		}
	}
	for _, id := range conShortIds {
		settingInfos = append(settingInfos, model.PackSettingInfo(settingModel[id]))
	}
	readIndexStart, err := model.GetReadIndexStart(ctx, conShortIds, userId)
	if err != nil {
		logrus.Errorf("[GetConversationSettings] GetReadIndexStart err. err = %v", err)
		return nil, err
	}
	readIndexEnd, err := model.GetReadIndexEnd(ctx, conShortIds, userId)
	if err != nil {
		logrus.Errorf("[GetConversationSettings] GetReadIndexEnd err. err = %v", err)
		return nil, err
	}
	readBadge, err := model.GetReadBadge(ctx, conShortIds, userId)
	if err != nil {
		logrus.Errorf("[GetConversationSettings] GetReadBadge err. err = %v", err)
		return nil, err
	}
	for i, conShortId := range conShortIds {
		if settingInfos[i] != nil {
			settingInfos[i].MinIndex = readIndexStart[conShortId]
			settingInfos[i].ReadIndexEnd = readIndexEnd[conShortId]
			settingInfos[i].ReadBadgeCount = readBadge[conShortId]
		}
	}
	return settingInfos, nil
	//userId,convIds
	//redis mget key；convId:userId,mysql,5小时
	//补齐，获取core，并发：判断是否成员，设置默认setting
	//获取readIndex、readBadge，abase mget，key；convId:userId
}

func fillSettingModel(ctx context.Context, userId int64, conShortIds []int64) (map[int64]*model.ConversationSettingInfo, error) {
	cores, err := GetConversationCores(ctx, conShortIds, false)
	if err != nil {
		logrus.Errorf("[fillSettingModel] GetConversationCores err. err = %v", err)
		return nil, err
	}
	settings := make([]*model.ConversationSettingInfo, 0)
	fillModel := make(map[int64]*model.ConversationSettingInfo)
	for _, core := range cores {
		setting := &model.ConversationSettingInfo{
			UserId:     userId,
			ConShortId: core.GetConShortId(),
			ConType:    core.GetConType(),
			ModifyTime: time.Now(),
			Extra:      core.GetExtra(),
		}
		settings = append(settings, setting)
		fillModel[setting.ConShortId] = setting
	}
	err = model.InsertSettingInfos(ctx, settings)
	if err != nil {
		logrus.Errorf("[fillSettingModel] InsertSettingInfos err. err = %v", err)
		return nil, err
	}
	return fillModel, nil
	//判断是否成员
}

func IncrConversationBadge(ctx context.Context, userId int64, conShortId int64) (int64, error) {
	key := fmt.Sprintf("badge:%d:%d", userId, conShortId)
	for i := 0; i < 3; i++ {
		badgeCnt, err := dal.KvrocksServer.Get(ctx, key)
		if errors.Is(err, redis.Nil) {
			newBadgeCntNum := int64(1)
			newBadgeCnt := "1"
			opt, err := dal.KvrocksServer.SetNX(ctx, key, newBadgeCnt)
			if err != nil {
				logrus.Errorf("[IncrConversationBadge] kvrocks SetNX err. err = %v", err)
				return 0, err
			}
			if opt {
				return newBadgeCntNum, nil
			}
		} else if err != nil {
			logrus.Errorf("[IncrConversationBadge] kvrocks Get err. err = %v", err)
			return 0, err
		} else {
			badgeCntNum, _ := strconv.ParseInt(badgeCnt, 10, 64)
			newBadgeCntNum := badgeCntNum + 1
			newBadgeCnt := strconv.FormatInt(newBadgeCntNum, 10)
			opt, err := dal.KvrocksServer.Cas(ctx, key, badgeCnt, newBadgeCnt)
			if err != nil {
				logrus.Errorf("[IncrConversationBadge] kvrocks Cas err. err = %v", err)
				return 0, err
			}
			if opt == 1 {
				return newBadgeCntNum, nil
			}
		}
		time.Sleep(SleepTime)
	}
	return 0, errors.New("retry too much")
}

func GetConversationBadges(ctx context.Context, userId int64, conShortIds []int64) ([]int64, error) {
	counts := make([]int64, 0)
	if len(conShortIds) == 0 {
		return counts, nil
	}
	keys := make([]string, 0)
	for _, id := range conShortIds {
		key := fmt.Sprintf("badge:%d:%d", userId, id)
		keys = append(keys, key)
	}
	countStrs, err := dal.KvrocksServer.MGet(ctx, keys)
	if err != nil {
		logrus.Errorf("[GetConversationBadge] kvrocks MGet err. err = %v", err)
		return nil, err
	}
	for _, countStr := range countStrs {
		if countStr == "" {
			counts = append(counts, 0)
		} else {
			count, _ := strconv.ParseInt(countStr, 10, 64)
			counts = append(counts, count)
		}
	}
	return counts, nil
}

func MarkRead(ctx context.Context, req *im.MarkReadRequest) (resp *im.MarkReadResponse, err error) {
	resp = &im.MarkReadResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	cores, err := GetConversationCores(ctx, []int64{req.GetConShortId()}, false)
	if err != nil {
		logrus.Errorf("[MarkRead] GetConversationCores err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, nil
	} else if len(cores) == 0 || cores[0] == nil || cores[0].GetStatus() != 0 {
		logrus.Errorf("[MarkRead] GetConversationCores err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Not_Found_Error, StatusMessage: "会话不存在"}
		return resp, nil
	}
	err = model.SetReadBadge(ctx, req.GetConShortId(), []int64{req.GetUserId()}, req.GetReadBadgeCount())
	if err != nil {
		logrus.Errorf("[MarkRead] SetReadBadge err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, nil
	}
	err = model.SetReadIndexEnd(ctx, req.GetConShortId(), []int64{req.GetUserId()}, req.GetReadConIndex())
	if err != nil {
		logrus.Errorf("[MarkRead] SetReadIndexEnd err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, nil
	}
	cmdMessage := map[string]interface{}{
		"read_index_end":   req.GetReadConIndex(),
		"read_badge_count": req.GetReadBadgeCount(),
	}
	cmdMessageByte, _ := json.Marshal(cmdMessage)
	sendMessageRequest := &im.SendMessageRequest{
		UserId:     req.GetUserId(),
		ConShortId: req.GetConShortId(),
		ConId:      cores[0].GetConId(),
		ConType:    cores[0].GetConType(),
		MsgType:    int32(im.MessageType_MarkRead),
		MsgContent: string(cmdMessageByte),
	}
	_, err = SendMessage(ctx, sendMessageRequest)
	if err != nil {
		logrus.Errorf("[MarkRead] SendMessage err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, nil
	}
	return resp, nil
	//设置总、各会话未读数(im_counter_manager_rust)->设置已读index和count(im_conversation_api,redis hash,abase xset)->
	//发送命令消息->更新最近会话(recent)->lastindex,保存,写命令链,push->发送已读消息(im_lastuser_api)
}
