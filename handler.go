package main

import (
	"context"
	"im/handler"
	"im/proto_gen/im"
)

type IMServerImpl struct {
	im.UnimplementedIMServiceServer
}

func (s *IMServerImpl) SendMessage(ctx context.Context, req *im.SendMessageRequest) (*im.SendMessageResponse, error) {
	return handler.SendMessage(ctx, req), nil
}

func (s *IMServerImpl) GetMessageByInit(ctx context.Context, req *im.GetMessageByInitRequest) (*im.GetMessageByInitResponse, error) {
	return handler.GetMessageByInit(ctx, req), nil
}

func (s *IMServerImpl) GetMessageByConversation(ctx context.Context, req *im.GetMessageByConversationRequest) (*im.GetMessageByConversationResponse, error) {
	return handler.GetMessageByConversation(ctx, req), nil
}

func (s *IMServerImpl) MarkRead(ctx context.Context, req *im.MarkReadRequest) (*im.MarkReadResponse, error) {
	return handler.MarkRead(ctx, req), nil
}

func (s *IMServerImpl) CreateConversation(ctx context.Context, req *im.CreateConversationRequest) (*im.CreateConversationResponse, error) {
	return &im.CreateConversationResponse{}, nil
}

func (s *IMServerImpl) GetConversationMembers(ctx context.Context, req *im.GetConversationMembersRequest) (*im.GetConversationMemberResponse, error) {
	return &im.GetConversationMemberResponse{}, nil
}

func (s *IMServerImpl) AddConversationMembers(ctx context.Context, req *im.AddConversationMembersRequest) (*im.AddConversationMembersResponse, error) {
	return &im.AddConversationMembersResponse{}, nil
}
