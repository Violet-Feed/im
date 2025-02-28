package conversation

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"im/dal"
	"im/handler/conversation/model"
	"im/handler/index"
	"im/handler/message"
	"im/proto_gen/im"
	"im/util"
	"math"
	"strconv"
)

const ConversationLimit = 100

func AddConversationMembers(ctx context.Context, req *im.AddConversationMembersRequest) (resp *im.AddConversationMembersResponse, err error) {
	//目前采用先操作再发消息的方案，可能会出现新成员会看到入群消息之前消息的问题
	resp = &im.AddConversationMembersResponse{
		BaseResp: &im.BaseResp{StatusCode: im.StatusCode_Success},
	}
	conShortId := req.GetConShortId()
	//判断群是否存在
	getConversationCoresRequest := &im.GetConversationCoresRequest{
		ConShortIds: []int64{conShortId},
	}
	getConversationCoresResponse, err := GetConversationCores(ctx, getConversationCoresRequest)
	if err != nil {
		logrus.Errorf("[AddConversationMembers] GetConversationCores err. err = %v", err)
		resp.BaseResp.StatusCode = im.StatusCode_Server_Error
		return resp, err
	}
	if len(getConversationCoresResponse.GetCoreInfos()) == 0 || getConversationCoresResponse.GetCoreInfos()[0].GetStatus() != 0 || getConversationCoresResponse.GetCoreInfos()[0].GetConType() != int32(im.ConversationType_Group_Chat) {
		resp.BaseResp.StatusCode = im.StatusCode_Not_Found_Error
		return resp, nil
	}
	//获取成员数量
	locked := dal.RedisServer.Lock(ctx, "user_count:"+strconv.FormatInt(conShortId, 10))
	if !locked {
		logrus.Errorf("[AddConversationMembers] Lock err.")
		resp.BaseResp.StatusCode = im.StatusCode_OverFrequency_Error
		return resp, err
	}
	defer dal.RedisServer.Unlock(ctx, "user_count:"+strconv.FormatInt(conShortId, 10))
	count, err := model.GetUserCount(ctx, conShortId)
	if err != nil {
		logrus.Errorf("[AddConversationMembers] GetUserCount err. err = %v", err)
		resp.BaseResp.StatusCode = im.StatusCode_Server_Error
		return resp, err
	}
	if count+len(req.GetMembers()) > ConversationLimit {
		resp.BaseResp.StatusCode = im.StatusCode_OverLimit_Error
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
		resp.BaseResp.StatusCode = im.StatusCode_Server_Error
		return resp, err
	}
	//获取设置index
	pullConversationIndexRequest := &im.PullConversationIndexRequest{
		ConShortId: util.Int64(conShortId),
		ConIndex:   util.Int64(math.MaxInt64),
		Limit:      util.Int64(1),
	}
	pullConversationIndexResponse, err := index.PullConversationIndex(ctx, pullConversationIndexRequest)
	if err != nil {
		logrus.Errorf("[AddConversationMembers] PullConversationIndex err. err = %v", err)
	} else {
		lastIndex := pullConversationIndexResponse.GetLastConIndex()
		err = model.SetReadIndexStart(ctx, conShortId, req.GetMembers(), lastIndex)
		if err != nil {
			logrus.Errorf("[AddConversationMembers] SetReadIndexStart err. err = %v", err)
		}
	}
	//发送进群命令消息
	cmdMessage := map[string]interface{}{
		"cmd_type":    im.SpecialCommandType_Add_Members,
		"operator":    req.GetOperator(),
		"add_members": req.GetMembers(),
	}
	cmdByte, _ := json.Marshal(cmdMessage)
	sendMessageRequest := &im.SendMessageRequest{
		UserId:     util.Int64(1),
		ConShortId: req.ConShortId,
		ConId:      req.ConId,
		ConType:    util.Int32(int32(im.ConversationType_Group_Chat)),
		MsgType:    util.Int32(int32(im.MessageType_SpecialCmd)),
		MsgContent: util.String(string(cmdByte)),
	}
	_, err = message.SendMessage(ctx, sendMessageRequest)
	if err != nil {
		logrus.Errorf("[AddConversationMembers] SendMessage err. err = %v", err)
	}
	//个人认为流程：redis加锁，获取成员数量判断是否超限，存入数据库，发送普通消息，解锁，入单链，对于新成员设置已读起点终点minIndex，入用户链，对于新成员获取badge设置readBadge
	//再次入群获取保存badgeCount->判断成员数量limit->保存userModels->再次判断limit，删除成员
	//拉会话链，获取index起点,设置已读起点->发送进群命令消息->发送挡板消息
	return resp, nil
}
