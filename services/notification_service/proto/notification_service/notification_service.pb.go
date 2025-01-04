// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.35.2
// 	protoc        v5.29.0
// source: proto/notification_service/notification_service.proto

package notification_service

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

type RegisterDeviceRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	UserID   uint32 `protobuf:"varint,1,opt,name=UserID,proto3" json:"UserID,omitempty"`
	FCMToken string `protobuf:"bytes,2,opt,name=FCMToken,proto3" json:"FCMToken,omitempty"`
}

func (x *RegisterDeviceRequest) Reset() {
	*x = RegisterDeviceRequest{}
	mi := &file_proto_notification_service_notification_service_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RegisterDeviceRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterDeviceRequest) ProtoMessage() {}

func (x *RegisterDeviceRequest) ProtoReflect() protoreflect.Message {
	mi := &file_proto_notification_service_notification_service_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterDeviceRequest.ProtoReflect.Descriptor instead.
func (*RegisterDeviceRequest) Descriptor() ([]byte, []int) {
	return file_proto_notification_service_notification_service_proto_rawDescGZIP(), []int{0}
}

func (x *RegisterDeviceRequest) GetUserID() uint32 {
	if x != nil {
		return x.UserID
	}
	return 0
}

func (x *RegisterDeviceRequest) GetFCMToken() string {
	if x != nil {
		return x.FCMToken
	}
	return ""
}

type RegisterDeviceResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool `protobuf:"varint,1,opt,name=Success,proto3" json:"Success,omitempty"`
}

func (x *RegisterDeviceResponse) Reset() {
	*x = RegisterDeviceResponse{}
	mi := &file_proto_notification_service_notification_service_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RegisterDeviceResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RegisterDeviceResponse) ProtoMessage() {}

func (x *RegisterDeviceResponse) ProtoReflect() protoreflect.Message {
	mi := &file_proto_notification_service_notification_service_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RegisterDeviceResponse.ProtoReflect.Descriptor instead.
func (*RegisterDeviceResponse) Descriptor() ([]byte, []int) {
	return file_proto_notification_service_notification_service_proto_rawDescGZIP(), []int{1}
}

func (x *RegisterDeviceResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

var File_proto_notification_service_notification_service_proto protoreflect.FileDescriptor

var file_proto_notification_service_notification_service_proto_rawDesc = []byte{
	0x0a, 0x35, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61,
	0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x6e, 0x6f, 0x74,
	0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0c, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63,
	0x61, 0x74, 0x69, 0x6f, 0x6e, 0x22, 0x4b, 0x0a, 0x15, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65,
	0x72, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x16,
	0x0a, 0x06, 0x55, 0x73, 0x65, 0x72, 0x49, 0x44, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x06,
	0x55, 0x73, 0x65, 0x72, 0x49, 0x44, 0x12, 0x1a, 0x0a, 0x08, 0x46, 0x43, 0x4d, 0x54, 0x6f, 0x6b,
	0x65, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x46, 0x43, 0x4d, 0x54, 0x6f, 0x6b,
	0x65, 0x6e, 0x22, 0x32, 0x0a, 0x16, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x44, 0x65,
	0x76, 0x69, 0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07,
	0x53, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x53,
	0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x32, 0x72, 0x0a, 0x13, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69,
	0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x5b, 0x0a,
	0x0e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x12,
	0x23, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x52,
	0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x1a, 0x24, 0x2e, 0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x2e, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x44, 0x65, 0x76, 0x69,
	0x63, 0x65, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x18, 0x5a, 0x16, 0x2e, 0x2f,
	0x6e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x73, 0x65, 0x72,
	0x76, 0x69, 0x63, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_proto_notification_service_notification_service_proto_rawDescOnce sync.Once
	file_proto_notification_service_notification_service_proto_rawDescData = file_proto_notification_service_notification_service_proto_rawDesc
)

func file_proto_notification_service_notification_service_proto_rawDescGZIP() []byte {
	file_proto_notification_service_notification_service_proto_rawDescOnce.Do(func() {
		file_proto_notification_service_notification_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_proto_notification_service_notification_service_proto_rawDescData)
	})
	return file_proto_notification_service_notification_service_proto_rawDescData
}

var file_proto_notification_service_notification_service_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_proto_notification_service_notification_service_proto_goTypes = []any{
	(*RegisterDeviceRequest)(nil),  // 0: notification.RegisterDeviceRequest
	(*RegisterDeviceResponse)(nil), // 1: notification.RegisterDeviceResponse
}
var file_proto_notification_service_notification_service_proto_depIdxs = []int32{
	0, // 0: notification.NotificationService.RegisterDevice:input_type -> notification.RegisterDeviceRequest
	1, // 1: notification.NotificationService.RegisterDevice:output_type -> notification.RegisterDeviceResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_proto_notification_service_notification_service_proto_init() }
func file_proto_notification_service_notification_service_proto_init() {
	if File_proto_notification_service_notification_service_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_proto_notification_service_notification_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_proto_notification_service_notification_service_proto_goTypes,
		DependencyIndexes: file_proto_notification_service_notification_service_proto_depIdxs,
		MessageInfos:      file_proto_notification_service_notification_service_proto_msgTypes,
	}.Build()
	File_proto_notification_service_notification_service_proto = out.File
	file_proto_notification_service_notification_service_proto_rawDesc = nil
	file_proto_notification_service_notification_service_proto_goTypes = nil
	file_proto_notification_service_notification_service_proto_depIdxs = nil
}
