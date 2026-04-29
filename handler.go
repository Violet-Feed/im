package main

import (
	"context"
	"im/biz"
	"im/proto_gen/im"
)

type IMServerImpl struct {
	im.UnimplementedIMServiceServer
}

func (s *IMServerImpl) GetInitInfo(ctx context.Context, req *im.GetInitInfoRequest) (*im.GetInitInfoResponse, error) {
	resp, _ := biz.GetInitInfo(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) SendMessage(ctx context.Context, req *im.SendMessageRequest) (*im.SendMessageResponse, error) {
	resp, _ := biz.SendMessage(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) GetMessageByUser(ctx context.Context, req *im.GetMessageByUserRequest) (*im.GetMessageByUserResponse, error) {
	resp, _ := biz.GetMessageByUser(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) GetCommandByUser(ctx context.Context, req *im.GetCommandByUserRequest) (*im.GetCommandByUserResponse, error) {
	resp, _ := biz.GetCommandByUser(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) GetMessageByConversation(ctx context.Context, req *im.GetMessageByConversationRequest) (*im.GetMessageByConversationResponse, error) {
	resp, _ := biz.GetMessageByConversation(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) MarkRead(ctx context.Context, req *im.MarkReadRequest) (*im.MarkReadResponse, error) {
	resp, _ := biz.MarkRead(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) RecallMessage(ctx context.Context, req *im.RecallMessageRequest) (*im.RecallMessageResponse, error) {
	resp, _ := biz.RecallMessage(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) CreateConversation(ctx context.Context, req *im.CreateConversationRequest) (*im.CreateConversationResponse, error) {
	resp, _ := biz.CreateConversation(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) GetConversationInfo(ctx context.Context, req *im.GetConversationInfoRequest) (*im.GetConversationInfoResponse, error) {
	resp, _ := biz.GetConversationInfo(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) UpdateConversationCore(ctx context.Context, req *im.UpdateConversationCoreRequest) (*im.UpdateConversationCoreResponse, error) {
	resp, _ := biz.UpdateConversationCore(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) UpdateConversationSetting(ctx context.Context, req *im.UpdateConversationSettingRequest) (*im.UpdateConversationSettingResponse, error) {
	resp, _ := biz.UpdateConversationSetting(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) UpdateConversationMember(ctx context.Context, req *im.UpdateConversationMemberRequest) (*im.UpdateConversationMemberResponse, error) {
	resp, _ := biz.UpdateConversationMember(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) GetConversationMembers(ctx context.Context, req *im.GetConversationMembersRequest) (*im.GetConversationMembersResponse, error) {
	resp, _ := biz.GetConversationMembers(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) GetConversationMembersByIds(ctx context.Context, req *im.GetConversationMembersByIdsRequest) (*im.GetConversationMembersByIdsResponse, error) {
	resp, _ := biz.GetConversationMembersByIds(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) AddConversationMembers(ctx context.Context, req *im.AddConversationMembersRequest) (*im.AddConversationMembersResponse, error) {
	resp, _ := biz.AddConversationMembers(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) RemoveConversationMember(ctx context.Context, req *im.RemoveConversationMemberRequest) (*im.RemoveConversationMemberResponse, error) {
	resp, _ := biz.RemoveConversationMember(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) GetMembersReadIndex(ctx context.Context, req *im.GetMembersReadIndexRequest) (*im.GetMembersReadIndexResponse, error) {
	resp, _ := biz.GetMembersReadIndex(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) SendNotice(ctx context.Context, req *im.SendNoticeRequest) (*im.SendNoticeResponse, error) {
	resp, _ := biz.SendNotice(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) GetNoticeList(ctx context.Context, req *im.GetNoticeListRequest) (*im.GetNoticeListResponse, error) {
	resp, _ := biz.GetNoticeList(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) GetNoticeAggList(ctx context.Context, req *im.GetNoticeAggListRequest) (*im.GetNoticeAggListResponse, error) {
	resp, _ := biz.GetNoticeAggList(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) GetNoticeCount(ctx context.Context, req *im.GetNoticeCountRequest) (*im.GetNoticeCountResponse, error) {
	resp, _ := biz.GetNoticeCount(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) MarkNoticeRead(ctx context.Context, req *im.MarkNoticeReadRequest) (*im.MarkNoticeReadResponse, error) {
	resp, _ := biz.MarkNoticeRead(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) AddConversationAgents(ctx context.Context, req *im.AddConversationAgentsRequest) (*im.AddConversationAgentsResponse, error) {
	resp, _ := biz.AddConversationAgents(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) RemoveConversationAgent(ctx context.Context, req *im.RemoveConversationAgentRequest) (*im.RemoveConversationAgentResponse, error) {
	resp, _ := biz.RemoveConversationAgent(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) GetConversationAgents(ctx context.Context, req *im.GetConversationAgentsRequest) (*im.GetConversationAgentsResponse, error) {
	resp, _ := biz.GetConversationAgents(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) GetConversationAgentsByIds(ctx context.Context, req *im.GetConversationAgentsByIdsRequest) (*im.GetConversationAgentsByIdsResponse, error) {
	resp, _ := biz.GetConversationAgentsByIds(ctx, req)
	return resp, nil
}
