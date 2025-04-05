package main

import (
	"context"
	"im/biz"
	"im/proto_gen/im"
)

type IMServerImpl struct {
	im.UnimplementedIMServiceServer
}

func (s *IMServerImpl) SendMessage(ctx context.Context, req *im.SendMessageRequest) (*im.SendMessageResponse, error) {
	resp, _ := biz.SendMessage(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) GetMessageByInit(ctx context.Context, req *im.GetMessageByInitRequest) (*im.GetMessageByInitResponse, error) {
	resp, _ := biz.GetMessageByInit(ctx, req)
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

func (s *IMServerImpl) CreateConversation(ctx context.Context, req *im.CreateConversationRequest) (*im.CreateConversationResponse, error) {
	resp, _ := biz.CreateConversation(ctx, req)
	return resp, nil
}

func (s *IMServerImpl) GetConversationMembers(ctx context.Context, req *im.GetConversationMembersRequest) (*im.GetConversationMemberResponse, error) {
	return &im.GetConversationMemberResponse{}, nil
}

func (s *IMServerImpl) AddConversationMembers(ctx context.Context, req *im.AddConversationMembersRequest) (*im.AddConversationMembersResponse, error) {
	resp, _ := biz.AddConversationMembers(ctx, req)
	return resp, nil
}
