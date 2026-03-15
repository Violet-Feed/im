package biz

import (
	"context"
	"encoding/json"
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
		logrus.Errorf("[AddConversationMembers] GetConversationCores err. err = %v", err)
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Server_Error}
		return resp, err
	}
	if len(cores) == 0 || cores[0] == nil || cores[0].GetStatus() != 0 || cores[0].GetConType() != int32(im.ConversationType_Group_Chat) {
		resp.BaseResp = &common.BaseResp{StatusCode: common.StatusCode_Not_Found_Error, StatusMessage: "会话不存在"}
		return resp, nil
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
		ConId:       req.GetConId(),
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
