// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v5.29.2
// source: proto/im.proto

package im

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

const (
	IMService_SendMessage_FullMethodName              = "/im.IMService/SendMessage"
	IMService_MessageGetByInit_FullMethodName         = "/im.IMService/MessageGetByInit"
	IMService_MessageGetByConversation_FullMethodName = "/im.IMService/MessageGetByConversation"
	IMService_MarkRead_FullMethodName                 = "/im.IMService/MarkRead"
	IMService_CreateConversation_FullMethodName       = "/im.IMService/CreateConversation"
	IMService_GetConversationMembers_FullMethodName   = "/im.IMService/GetConversationMembers"
	IMService_AddConversationMembers_FullMethodName   = "/im.IMService/AddConversationMembers"
)

// IMServiceClient is the client API for IMService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type IMServiceClient interface {
	SendMessage(ctx context.Context, in *SendMessageRequest, opts ...grpc.CallOption) (*SendMessageResponse, error)
	MessageGetByInit(ctx context.Context, in *MessageGetByInitRequest, opts ...grpc.CallOption) (*MessageGetByInitResponse, error)
	MessageGetByConversation(ctx context.Context, in *MessageGetByConversationRequest, opts ...grpc.CallOption) (*MessageGetByConversationResponse, error)
	MarkRead(ctx context.Context, in *MarkReadRequest, opts ...grpc.CallOption) (*MarkReadResponse, error)
	CreateConversation(ctx context.Context, in *CreateConversationRequest, opts ...grpc.CallOption) (*CreateConversationResponse, error)
	GetConversationMembers(ctx context.Context, in *GetConversationMembersRequest, opts ...grpc.CallOption) (*GetConversationMemberResponse, error)
	AddConversationMembers(ctx context.Context, in *AddConversationMembersRequest, opts ...grpc.CallOption) (*AddConversationMembersResponse, error)
}

type iMServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewIMServiceClient(cc grpc.ClientConnInterface) IMServiceClient {
	return &iMServiceClient{cc}
}

func (c *iMServiceClient) SendMessage(ctx context.Context, in *SendMessageRequest, opts ...grpc.CallOption) (*SendMessageResponse, error) {
	out := new(SendMessageResponse)
	err := c.cc.Invoke(ctx, IMService_SendMessage_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iMServiceClient) MessageGetByInit(ctx context.Context, in *MessageGetByInitRequest, opts ...grpc.CallOption) (*MessageGetByInitResponse, error) {
	out := new(MessageGetByInitResponse)
	err := c.cc.Invoke(ctx, IMService_MessageGetByInit_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iMServiceClient) MessageGetByConversation(ctx context.Context, in *MessageGetByConversationRequest, opts ...grpc.CallOption) (*MessageGetByConversationResponse, error) {
	out := new(MessageGetByConversationResponse)
	err := c.cc.Invoke(ctx, IMService_MessageGetByConversation_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iMServiceClient) MarkRead(ctx context.Context, in *MarkReadRequest, opts ...grpc.CallOption) (*MarkReadResponse, error) {
	out := new(MarkReadResponse)
	err := c.cc.Invoke(ctx, IMService_MarkRead_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iMServiceClient) CreateConversation(ctx context.Context, in *CreateConversationRequest, opts ...grpc.CallOption) (*CreateConversationResponse, error) {
	out := new(CreateConversationResponse)
	err := c.cc.Invoke(ctx, IMService_CreateConversation_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iMServiceClient) GetConversationMembers(ctx context.Context, in *GetConversationMembersRequest, opts ...grpc.CallOption) (*GetConversationMemberResponse, error) {
	out := new(GetConversationMemberResponse)
	err := c.cc.Invoke(ctx, IMService_GetConversationMembers_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *iMServiceClient) AddConversationMembers(ctx context.Context, in *AddConversationMembersRequest, opts ...grpc.CallOption) (*AddConversationMembersResponse, error) {
	out := new(AddConversationMembersResponse)
	err := c.cc.Invoke(ctx, IMService_AddConversationMembers_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// IMServiceServer is the server API for IMService service.
// All implementations must embed UnimplementedIMServiceServer
// for forward compatibility
type IMServiceServer interface {
	SendMessage(context.Context, *SendMessageRequest) (*SendMessageResponse, error)
	MessageGetByInit(context.Context, *MessageGetByInitRequest) (*MessageGetByInitResponse, error)
	MessageGetByConversation(context.Context, *MessageGetByConversationRequest) (*MessageGetByConversationResponse, error)
	MarkRead(context.Context, *MarkReadRequest) (*MarkReadResponse, error)
	CreateConversation(context.Context, *CreateConversationRequest) (*CreateConversationResponse, error)
	GetConversationMembers(context.Context, *GetConversationMembersRequest) (*GetConversationMemberResponse, error)
	AddConversationMembers(context.Context, *AddConversationMembersRequest) (*AddConversationMembersResponse, error)
	mustEmbedUnimplementedIMServiceServer()
}

// UnimplementedIMServiceServer must be embedded to have forward compatible implementations.
type UnimplementedIMServiceServer struct {
}

func (UnimplementedIMServiceServer) SendMessage(context.Context, *SendMessageRequest) (*SendMessageResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMessage not implemented")
}
func (UnimplementedIMServiceServer) MessageGetByInit(context.Context, *MessageGetByInitRequest) (*MessageGetByInitResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MessageGetByInit not implemented")
}
func (UnimplementedIMServiceServer) MessageGetByConversation(context.Context, *MessageGetByConversationRequest) (*MessageGetByConversationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MessageGetByConversation not implemented")
}
func (UnimplementedIMServiceServer) MarkRead(context.Context, *MarkReadRequest) (*MarkReadResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MarkRead not implemented")
}
func (UnimplementedIMServiceServer) CreateConversation(context.Context, *CreateConversationRequest) (*CreateConversationResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CreateConversation not implemented")
}
func (UnimplementedIMServiceServer) GetConversationMembers(context.Context, *GetConversationMembersRequest) (*GetConversationMemberResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetConversationMembers not implemented")
}
func (UnimplementedIMServiceServer) AddConversationMembers(context.Context, *AddConversationMembersRequest) (*AddConversationMembersResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method AddConversationMembers not implemented")
}
func (UnimplementedIMServiceServer) mustEmbedUnimplementedIMServiceServer() {}

// UnsafeIMServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to IMServiceServer will
// result in compilation errors.
type UnsafeIMServiceServer interface {
	mustEmbedUnimplementedIMServiceServer()
}

func RegisterIMServiceServer(s grpc.ServiceRegistrar, srv IMServiceServer) {
	s.RegisterService(&IMService_ServiceDesc, srv)
}

func _IMService_SendMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SendMessageRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IMServiceServer).SendMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IMService_SendMessage_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IMServiceServer).SendMessage(ctx, req.(*SendMessageRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IMService_MessageGetByInit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MessageGetByInitRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IMServiceServer).MessageGetByInit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IMService_MessageGetByInit_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IMServiceServer).MessageGetByInit(ctx, req.(*MessageGetByInitRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IMService_MessageGetByConversation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MessageGetByConversationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IMServiceServer).MessageGetByConversation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IMService_MessageGetByConversation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IMServiceServer).MessageGetByConversation(ctx, req.(*MessageGetByConversationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IMService_MarkRead_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MarkReadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IMServiceServer).MarkRead(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IMService_MarkRead_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IMServiceServer).MarkRead(ctx, req.(*MarkReadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IMService_CreateConversation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CreateConversationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IMServiceServer).CreateConversation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IMService_CreateConversation_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IMServiceServer).CreateConversation(ctx, req.(*CreateConversationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IMService_GetConversationMembers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetConversationMembersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IMServiceServer).GetConversationMembers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IMService_GetConversationMembers_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IMServiceServer).GetConversationMembers(ctx, req.(*GetConversationMembersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _IMService_AddConversationMembers_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(AddConversationMembersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(IMServiceServer).AddConversationMembers(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: IMService_AddConversationMembers_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(IMServiceServer).AddConversationMembers(ctx, req.(*AddConversationMembersRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// IMService_ServiceDesc is the grpc.ServiceDesc for IMService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var IMService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "im.IMService",
	HandlerType: (*IMServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendMessage",
			Handler:    _IMService_SendMessage_Handler,
		},
		{
			MethodName: "MessageGetByInit",
			Handler:    _IMService_MessageGetByInit_Handler,
		},
		{
			MethodName: "MessageGetByConversation",
			Handler:    _IMService_MessageGetByConversation_Handler,
		},
		{
			MethodName: "MarkRead",
			Handler:    _IMService_MarkRead_Handler,
		},
		{
			MethodName: "CreateConversation",
			Handler:    _IMService_CreateConversation_Handler,
		},
		{
			MethodName: "GetConversationMembers",
			Handler:    _IMService_GetConversationMembers_Handler,
		},
		{
			MethodName: "AddConversationMembers",
			Handler:    _IMService_AddConversationMembers_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/im.proto",
}
