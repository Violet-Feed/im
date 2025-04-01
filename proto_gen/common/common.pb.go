// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        v5.29.2
// source: proto/common.proto

package common

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type StatusCode int32

const (
	StatusCode_StatusCode_Not_Use  StatusCode = 0
	StatusCode_Success             StatusCode = 1000
	StatusCode_Server_Error        StatusCode = 1001
	StatusCode_Param_Error         StatusCode = 1002
	StatusCode_OverFrequency_Error StatusCode = 1003
	StatusCode_OverLimit_Error     StatusCode = 1004
	StatusCode_Duplicate_Error     StatusCode = 1005
	StatusCode_RetryTime_Error     StatusCode = 1006
	StatusCode_Not_Found_Error     StatusCode = 1007
	StatusCode_Unknown_Error       StatusCode = 9999
)

// Enum value maps for StatusCode.
var (
	StatusCode_name = map[int32]string{
		0:    "StatusCode_Not_Use",
		1000: "Success",
		1001: "Server_Error",
		1002: "Param_Error",
		1003: "OverFrequency_Error",
		1004: "OverLimit_Error",
		1005: "Duplicate_Error",
		1006: "RetryTime_Error",
		1007: "Not_Found_Error",
		9999: "Unknown_Error",
	}
	StatusCode_value = map[string]int32{
		"StatusCode_Not_Use":  0,
		"Success":             1000,
		"Server_Error":        1001,
		"Param_Error":         1002,
		"OverFrequency_Error": 1003,
		"OverLimit_Error":     1004,
		"Duplicate_Error":     1005,
		"RetryTime_Error":     1006,
		"Not_Found_Error":     1007,
		"Unknown_Error":       9999,
	}
)

func (x StatusCode) Enum() *StatusCode {
	p := new(StatusCode)
	*p = x
	return p
}

func (x StatusCode) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (StatusCode) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_common_proto_enumTypes[0].Descriptor()
}

func (StatusCode) Type() protoreflect.EnumType {
	return &file_proto_common_proto_enumTypes[0]
}

func (x StatusCode) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use StatusCode.Descriptor instead.
func (StatusCode) EnumDescriptor() ([]byte, []int) {
	return file_proto_common_proto_rawDescGZIP(), []int{0}
}

type SpecialUser int32

const (
	SpecialUser_SpecialUser_Not_Use SpecialUser = 0
	SpecialUser_System              SpecialUser = 1
	SpecialUser_Action              SpecialUser = 2
	SpecialUser_Market              SpecialUser = 3
	SpecialUser_AI                  SpecialUser = 4
	SpecialUser_Conversation        SpecialUser = 5
)

// Enum value maps for SpecialUser.
var (
	SpecialUser_name = map[int32]string{
		0: "SpecialUser_Not_Use",
		1: "System",
		2: "Action",
		3: "Market",
		4: "AI",
		5: "Conversation",
	}
	SpecialUser_value = map[string]int32{
		"SpecialUser_Not_Use": 0,
		"System":              1,
		"Action":              2,
		"Market":              3,
		"AI":                  4,
		"Conversation":        5,
	}
)

func (x SpecialUser) Enum() *SpecialUser {
	p := new(SpecialUser)
	*p = x
	return p
}

func (x SpecialUser) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SpecialUser) Descriptor() protoreflect.EnumDescriptor {
	return file_proto_common_proto_enumTypes[1].Descriptor()
}

func (SpecialUser) Type() protoreflect.EnumType {
	return &file_proto_common_proto_enumTypes[1]
}

func (x SpecialUser) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SpecialUser.Descriptor instead.
func (SpecialUser) EnumDescriptor() ([]byte, []int) {
	return file_proto_common_proto_rawDescGZIP(), []int{1}
}

type BaseResp struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	StatusCode    StatusCode `protobuf:"varint,1,opt,name=StatusCode,proto3,enum=common.StatusCode" json:"StatusCode,omitempty"`
	StatusMessage string     `protobuf:"bytes,2,opt,name=StatusMessage,proto3" json:"StatusMessage,omitempty"`
}

func (x *BaseResp) Reset() {
	*x = BaseResp{}
	if protoimpl.UnsafeEnabled {
		mi := &file_proto_common_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BaseResp) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BaseResp) ProtoMessage() {}

func (x *BaseResp) ProtoReflect() protoreflect.Message {
	mi := &file_proto_common_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BaseResp.ProtoReflect.Descriptor instead.
func (*BaseResp) Descriptor() ([]byte, []int) {
	return file_proto_common_proto_rawDescGZIP(), []int{0}
}

func (x *BaseResp) GetStatusCode() StatusCode {
	if x != nil {
		return x.StatusCode
	}
	return StatusCode_StatusCode_Not_Use
}

func (x *BaseResp) GetStatusMessage() string {
	if x != nil {
		return x.StatusMessage
	}
	return ""
}

var File_proto_common_proto protoreflect.FileDescriptor

var file_proto_common_proto_rawDesc = []byte{
	0x0a, 0x12, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x22, 0x64, 0x0a, 0x08,
	0x42, 0x61, 0x73, 0x65, 0x52, 0x65, 0x73, 0x70, 0x12, 0x32, 0x0a, 0x0a, 0x53, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x43, 0x6f, 0x64, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x12, 0x2e, 0x63,
	0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x43, 0x6f, 0x64, 0x65,
	0x52, 0x0a, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x43, 0x6f, 0x64, 0x65, 0x12, 0x24, 0x0a, 0x0d,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0d, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x2a, 0xdd, 0x01, 0x0a, 0x0a, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x43, 0x6f, 0x64,
	0x65, 0x12, 0x16, 0x0a, 0x12, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x43, 0x6f, 0x64, 0x65, 0x5f,
	0x4e, 0x6f, 0x74, 0x5f, 0x55, 0x73, 0x65, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x07, 0x53, 0x75, 0x63,
	0x63, 0x65, 0x73, 0x73, 0x10, 0xe8, 0x07, 0x12, 0x11, 0x0a, 0x0c, 0x53, 0x65, 0x72, 0x76, 0x65,
	0x72, 0x5f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x10, 0xe9, 0x07, 0x12, 0x10, 0x0a, 0x0b, 0x50, 0x61,
	0x72, 0x61, 0x6d, 0x5f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x10, 0xea, 0x07, 0x12, 0x18, 0x0a, 0x13,
	0x4f, 0x76, 0x65, 0x72, 0x46, 0x72, 0x65, 0x71, 0x75, 0x65, 0x6e, 0x63, 0x79, 0x5f, 0x45, 0x72,
	0x72, 0x6f, 0x72, 0x10, 0xeb, 0x07, 0x12, 0x14, 0x0a, 0x0f, 0x4f, 0x76, 0x65, 0x72, 0x4c, 0x69,
	0x6d, 0x69, 0x74, 0x5f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x10, 0xec, 0x07, 0x12, 0x14, 0x0a, 0x0f,
	0x44, 0x75, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x65, 0x5f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x10,
	0xed, 0x07, 0x12, 0x14, 0x0a, 0x0f, 0x52, 0x65, 0x74, 0x72, 0x79, 0x54, 0x69, 0x6d, 0x65, 0x5f,
	0x45, 0x72, 0x72, 0x6f, 0x72, 0x10, 0xee, 0x07, 0x12, 0x14, 0x0a, 0x0f, 0x4e, 0x6f, 0x74, 0x5f,
	0x46, 0x6f, 0x75, 0x6e, 0x64, 0x5f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x10, 0xef, 0x07, 0x12, 0x12,
	0x0a, 0x0d, 0x55, 0x6e, 0x6b, 0x6e, 0x6f, 0x77, 0x6e, 0x5f, 0x45, 0x72, 0x72, 0x6f, 0x72, 0x10,
	0x8f, 0x4e, 0x2a, 0x64, 0x0a, 0x0b, 0x53, 0x70, 0x65, 0x63, 0x69, 0x61, 0x6c, 0x55, 0x73, 0x65,
	0x72, 0x12, 0x17, 0x0a, 0x13, 0x53, 0x70, 0x65, 0x63, 0x69, 0x61, 0x6c, 0x55, 0x73, 0x65, 0x72,
	0x5f, 0x4e, 0x6f, 0x74, 0x5f, 0x55, 0x73, 0x65, 0x10, 0x00, 0x12, 0x0a, 0x0a, 0x06, 0x53, 0x79,
	0x73, 0x74, 0x65, 0x6d, 0x10, 0x01, 0x12, 0x0a, 0x0a, 0x06, 0x41, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x10, 0x02, 0x12, 0x0a, 0x0a, 0x06, 0x4d, 0x61, 0x72, 0x6b, 0x65, 0x74, 0x10, 0x03, 0x12, 0x06,
	0x0a, 0x02, 0x41, 0x49, 0x10, 0x04, 0x12, 0x10, 0x0a, 0x0c, 0x43, 0x6f, 0x6e, 0x76, 0x65, 0x72,
	0x73, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x10, 0x05, 0x42, 0x3d, 0x0a, 0x26, 0x76, 0x69, 0x6f, 0x6c,
	0x65, 0x74, 0x2e, 0x67, 0x61, 0x74, 0x65, 0x77, 0x61, 0x79, 0x2e, 0x63, 0x6f, 0x6d, 0x6d, 0x6f,
	0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x5f, 0x67, 0x65, 0x6e, 0x2e, 0x63, 0x6f, 0x6d, 0x6d,
	0x6f, 0x6e, 0x50, 0x01, 0x5a, 0x11, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x5f, 0x67, 0x65, 0x6e,
	0x2f, 0x63, 0x6f, 0x6d, 0x6d, 0x6f, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_common_proto_rawDescOnce sync.Once
	file_proto_common_proto_rawDescData = file_proto_common_proto_rawDesc
)

func file_proto_common_proto_rawDescGZIP() []byte {
	file_proto_common_proto_rawDescOnce.Do(func() {
		file_proto_common_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_common_proto_rawDescData)
	})
	return file_proto_common_proto_rawDescData
}

var file_proto_common_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_proto_common_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_proto_common_proto_goTypes = []interface{}{
	(StatusCode)(0),  // 0: common.StatusCode
	(SpecialUser)(0), // 1: common.SpecialUser
	(*BaseResp)(nil), // 2: common.BaseResp
}
var file_proto_common_proto_depIdxs = []int32{
	0, // 0: common.BaseResp.StatusCode:type_name -> common.StatusCode
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_proto_common_proto_init() }
func file_proto_common_proto_init() {
	if File_proto_common_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_proto_common_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BaseResp); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_common_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_proto_common_proto_goTypes,
		DependencyIndexes: file_proto_common_proto_depIdxs,
		EnumInfos:         file_proto_common_proto_enumTypes,
		MessageInfos:      file_proto_common_proto_msgTypes,
	}.Build()
	File_proto_common_proto = out.File
	file_proto_common_proto_rawDesc = nil
	file_proto_common_proto_goTypes = nil
	file_proto_common_proto_depIdxs = nil
}
