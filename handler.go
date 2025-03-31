package main

import (
	"context"
	"im/proto_gen/im"
)

type IMServerImpl struct {
	im.UnimplementedIMServiceServer
}

func (s *IMServerImpl) SendMessage(ctx context.Context, req *im.SendMessageRequest) (*im.SendMessageResponse, error) {
	return &im.SendMessageResponse{}, nil
}

func (s *IMServerImpl) MessageGetByInit(ctx context.Context, req *im.MessageGetByInitRequest) (*im.MessageGetByInitResponse, error) {
	return &im.MessageGetByInitResponse{}, nil
}

func (s *IMServerImpl) MessageGetByConversation(ctx context.Context, req *im.MessageGetByConversationRequest) (*im.MessageGetByConversationResponse, error) {
	return &im.MessageGetByConversationResponse{}, nil
}

func (s *IMServerImpl) MarkRead(ctx context.Context, req *im.MarkReadRequest) (*im.MarkReadResponse, error) {
	return &im.MarkReadResponse{}, nil
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
