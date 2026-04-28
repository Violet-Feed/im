package biz

import (
	"context"
	"encoding/json"
	"errors"
	"im/biz/model"
	"im/proto_gen/common"
	"im/proto_gen/im"

	"github.com/sirupsen/logrus"
)

func AddConversationAgents(ctx context.Context, req *im.AddConversationAgentsRequest) (*im.AddConversationAgentsResponse, error) {
	resp := &im.AddConversationAgentsResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	conShortId := req.GetConShortId()
	//判断群是否存在
	cores, err := GetConversationCores(ctx, []int64{conShortId}, false)
	if err != nil {
		logrus.Errorf("[AddConversationAgents] GetConversationCores err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, err
	}
	if len(cores) == 0 || cores[0] == nil || cores[0].GetStatus() != 0 || cores[0].GetConType() != int32(im.ConversationType_Group_Chat) {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Not_Found_Error, StatusMessage: "conversation not found"}
		return resp, nil
	}
	core := cores[0]
	isMember, err := checkConversationMember(ctx, req.GetConShortId(), core.GetConId(), core.GetConType(), int32(im.SenderType_User), req.GetOperator())
	if err != nil {
		logrus.Errorf("[AddConversationAgents] checkConversationMember err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, err
	}
	if !isMember {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Auth_Error, StatusMessage: "not conversation member"}
		return resp, errors.New("not conversation member")
	}
	//todo:获取agent数量
	//创建agent
	var agentModels []*model.ConversationAgentInfo
	for _, agentId := range req.AgentIds {
		agentModels = append(agentModels, model.PackAgentModel(agentId, req))
	}
	if err := model.InsertAgentInfos(ctx, agentModels); err != nil {
		logrus.Errorf("[AddConversationAgents] InsertAgentInfos err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, err
	}
	//发送消息
	conMessage := map[string]interface{}{
		"type":     im.ConMessageType_Add_Agent,
		"operator": req.GetOperator(),
		"content":  req.GetAgentIds(),
	}
	conMessageByte, _ := json.Marshal(conMessage)
	sendMessageRequest := &im.SendMessageRequest{
		SenderId:    0,
		SenderType:  int32(im.SenderType_Conv),
		ConShortId:  req.GetConShortId(),
		ConId:       core.GetConId(),
		ConType:     int32(im.ConversationType_Group_Chat),
		MsgType:     int32(im.MessageType_Conversation),
		MsgContent:  string(conMessageByte),
		ClientMsgId: 0,
	}
	logrus.Infof("[AddConversationAgents] sendMessageRequest = %v", sendMessageRequest)
	_, err = SendMessage(ctx, sendMessageRequest)
	if err != nil {
		logrus.Errorf("[AddConversationAgents] SendMessage err. err = %v", err)
	}
	return resp, nil
}

func RemoveConversationAgent(ctx context.Context, req *im.RemoveConversationAgentRequest) (*im.RemoveConversationAgentResponse, error) {
	resp := &im.RemoveConversationAgentResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	cores, err := GetConversationCores(ctx, []int64{req.GetConShortId()}, false)
	if err != nil {
		logrus.Errorf("[RemoveConversationAgents] GetConversationCores err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, err
	}
	if len(cores) == 0 || cores[0] == nil || cores[0].GetStatus() != 0 || cores[0].GetConType() != int32(im.ConversationType_Group_Chat) {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Not_Found_Error, StatusMessage: "conversation not found"}
		return resp, nil
	}
	core := cores[0]
	isMember, err := checkConversationMember(ctx, req.GetConShortId(), core.GetConId(), core.GetConType(), int32(im.SenderType_User), req.GetOperator())
	if err != nil {
		logrus.Errorf("[RemoveConversationAgents] checkConversationMember err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, err
	}
	if !isMember {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Auth_Error, StatusMessage: "not conversation member"}
		return resp, errors.New("not conversation member")
	}
	if err := model.DeleteAgentInfo(ctx, req.GetConShortId(), req.GetAgentId()); err != nil {
		logrus.Errorf("[RemoveConversationAgents] DeleteAgentInfos err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, err
	}
	conMessage := map[string]interface{}{
		"type":     im.ConMessageType_Remove_Agent,
		"operator": req.GetOperator(),
		"content":  req.GetAgentId(),
	}
	conMessageByte, _ := json.Marshal(conMessage)
	sendMessageRequest := &im.SendMessageRequest{
		SenderId:    0,
		SenderType:  int32(im.SenderType_Conv),
		ConShortId:  req.GetConShortId(),
		ConId:       core.GetConId(),
		ConType:     int32(im.ConversationType_Group_Chat),
		MsgType:     int32(im.MessageType_Conversation),
		MsgContent:  string(conMessageByte),
		ClientMsgId: 0,
	}
	_, err = SendMessage(ctx, sendMessageRequest)
	if err != nil {
		logrus.Errorf("[RemoveConversationAgents] SendMessage err. err = %v", err)
	}
	return resp, nil
}

func GetConversationAgents(ctx context.Context, req *im.GetConversationAgentsRequest) (*im.GetConversationAgentsResponse, error) {
	resp := &im.GetConversationAgentsResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	agents, err := model.GetAgentInfos(ctx, req.ConShortId)
	if err != nil {
		logrus.Errorf("[GetConversationAgents] GetAgentInfos err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, nil
	}
	var agentInfos []*im.ConversationAgentInfo
	for _, agent := range agents {
		agentInfos = append(agentInfos, model.PackAgentInfo(agent))
	}
	resp.Agents = agentInfos
	return resp, nil
}

func GetConversationAgentsByIds(ctx context.Context, req *im.GetConversationAgentsByIdsRequest) (*im.GetConversationAgentsByIdsResponse, error) {
	resp := &im.GetConversationAgentsByIdsResponse{
		BaseResp: &common.BaseResp{StatusCode: common.StatusCode_Success},
	}
	return resp, nil
}
