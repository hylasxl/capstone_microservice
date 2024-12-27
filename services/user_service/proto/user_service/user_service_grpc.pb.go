// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.0
// source: proto/user_service/user_service.proto

package user_service

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.64.0 or later.
const _ = grpc.SupportPackageIsVersion9

const (
	UserService_Login_FullMethodName                     = "/user.UserService/Login"
	UserService_Signup_FullMethodName                    = "/user.UserService/Signup"
	UserService_CheckExistingUsername_FullMethodName     = "/user.UserService/CheckExistingUsername"
	UserService_CheckExistingEmail_FullMethodName        = "/user.UserService/CheckExistingEmail"
	UserService_CheckExistingPhone_FullMethodName        = "/user.UserService/CheckExistingPhone"
	UserService_CheckValidUser_FullMethodName            = "/user.UserService/CheckValidUser"
	UserService_GetListAccountDisplayInfo_FullMethodName = "/user.UserService/GetListAccountDisplayInfo"
	UserService_GetAccountInfo_FullMethodName            = "/user.UserService/GetAccountInfo"
)

// UserServiceClient is the client API for UserService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UserServiceClient interface {
	Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error)
	Signup(ctx context.Context, in *SignupRequest, opts ...grpc.CallOption) (*SignupResponse, error)
	CheckExistingUsername(ctx context.Context, in *CheckExistingUsernameRequest, opts ...grpc.CallOption) (*CheckExistingUsernameResponse, error)
	CheckExistingEmail(ctx context.Context, in *CheckExistingEmailRequest, opts ...grpc.CallOption) (*CheckExistingEmailResponse, error)
	CheckExistingPhone(ctx context.Context, in *CheckExistingPhoneRequest, opts ...grpc.CallOption) (*CheckExistingPhoneResponse, error)
	CheckValidUser(ctx context.Context, in *CheckValidUserRequest, opts ...grpc.CallOption) (*CheckValidUserResponse, error)
	GetListAccountDisplayInfo(ctx context.Context, in *GetListAccountDisplayInfoRequest, opts ...grpc.CallOption) (*GetListAccountDisplayInfoResponse, error)
	GetAccountInfo(ctx context.Context, in *GetAccountInfoRequest, opts ...grpc.CallOption) (*GetAccountInfoResponse, error)
}

type userServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewUserServiceClient(cc grpc.ClientConnInterface) UserServiceClient {
	return &userServiceClient{cc}
}

func (c *userServiceClient) Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LoginResponse)
	err := c.cc.Invoke(ctx, UserService_Login_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) Signup(ctx context.Context, in *SignupRequest, opts ...grpc.CallOption) (*SignupResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(SignupResponse)
	err := c.cc.Invoke(ctx, UserService_Signup_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) CheckExistingUsername(ctx context.Context, in *CheckExistingUsernameRequest, opts ...grpc.CallOption) (*CheckExistingUsernameResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CheckExistingUsernameResponse)
	err := c.cc.Invoke(ctx, UserService_CheckExistingUsername_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) CheckExistingEmail(ctx context.Context, in *CheckExistingEmailRequest, opts ...grpc.CallOption) (*CheckExistingEmailResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CheckExistingEmailResponse)
	err := c.cc.Invoke(ctx, UserService_CheckExistingEmail_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) CheckExistingPhone(ctx context.Context, in *CheckExistingPhoneRequest, opts ...grpc.CallOption) (*CheckExistingPhoneResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CheckExistingPhoneResponse)
	err := c.cc.Invoke(ctx, UserService_CheckExistingPhone_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) CheckValidUser(ctx context.Context, in *CheckValidUserRequest, opts ...grpc.CallOption) (*CheckValidUserResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(CheckValidUserResponse)
	err := c.cc.Invoke(ctx, UserService_CheckValidUser_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) GetListAccountDisplayInfo(ctx context.Context, in *GetListAccountDisplayInfoRequest, opts ...grpc.CallOption) (*GetListAccountDisplayInfoResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetListAccountDisplayInfoResponse)
	err := c.cc.Invoke(ctx, UserService_GetListAccountDisplayInfo_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) GetAccountInfo(ctx context.Context, in *GetAccountInfoRequest, opts ...grpc.CallOption) (*GetAccountInfoResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(GetAccountInfoResponse)
	err := c.cc.Invoke(ctx, UserService_GetAccountInfo_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UserServiceServer is the server API for UserService service.
// All implementations must embed UnimplementedUserServiceServer
// for forward compatibility.
type UserServiceServer interface {
	Login(context.Context, *LoginRequest) (*LoginResponse, error)
	Signup(context.Context, *SignupRequest) (*SignupResponse, error)
	CheckExistingUsername(context.Context, *CheckExistingUsernameRequest) (*CheckExistingUsernameResponse, error)
	CheckExistingEmail(context.Context, *CheckExistingEmailRequest) (*CheckExistingEmailResponse, error)
	CheckExistingPhone(context.Context, *CheckExistingPhoneRequest) (*CheckExistingPhoneResponse, error)
	CheckValidUser(context.Context, *CheckValidUserRequest) (*CheckValidUserResponse, error)
	GetListAccountDisplayInfo(context.Context, *GetListAccountDisplayInfoRequest) (*GetListAccountDisplayInfoResponse, error)
	GetAccountInfo(context.Context, *GetAccountInfoRequest) (*GetAccountInfoResponse, error)
	mustEmbedUnimplementedUserServiceServer()
}

// UnimplementedUserServiceServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedUserServiceServer struct{}

func (UnimplementedUserServiceServer) Login(context.Context, *LoginRequest) (*LoginResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Login not implemented")
}
func (UnimplementedUserServiceServer) Signup(context.Context, *SignupRequest) (*SignupResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Signup not implemented")
}
func (UnimplementedUserServiceServer) CheckExistingUsername(context.Context, *CheckExistingUsernameRequest) (*CheckExistingUsernameResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckExistingUsername not implemented")
}
func (UnimplementedUserServiceServer) CheckExistingEmail(context.Context, *CheckExistingEmailRequest) (*CheckExistingEmailResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckExistingEmail not implemented")
}
func (UnimplementedUserServiceServer) CheckExistingPhone(context.Context, *CheckExistingPhoneRequest) (*CheckExistingPhoneResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckExistingPhone not implemented")
}
func (UnimplementedUserServiceServer) CheckValidUser(context.Context, *CheckValidUserRequest) (*CheckValidUserResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CheckValidUser not implemented")
}
func (UnimplementedUserServiceServer) GetListAccountDisplayInfo(context.Context, *GetListAccountDisplayInfoRequest) (*GetListAccountDisplayInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetListAccountDisplayInfo not implemented")
}
func (UnimplementedUserServiceServer) GetAccountInfo(context.Context, *GetAccountInfoRequest) (*GetAccountInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetAccountInfo not implemented")
}
func (UnimplementedUserServiceServer) mustEmbedUnimplementedUserServiceServer() {}
func (UnimplementedUserServiceServer) testEmbeddedByValue()                     {}

// UnsafeUserServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UserServiceServer will
// result in compilation errors.
type UnsafeUserServiceServer interface {
	mustEmbedUnimplementedUserServiceServer()
}

func RegisterUserServiceServer(s grpc.ServiceRegistrar, srv UserServiceServer) {
	// If the following call pancis, it indicates UnimplementedUserServiceServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&UserService_ServiceDesc, srv)
}

func _UserService_Login_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LoginRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).Login(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_Login_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).Login(ctx, req.(*LoginRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_Signup_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SignupRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).Signup(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_Signup_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).Signup(ctx, req.(*SignupRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_CheckExistingUsername_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckExistingUsernameRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).CheckExistingUsername(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_CheckExistingUsername_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).CheckExistingUsername(ctx, req.(*CheckExistingUsernameRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_CheckExistingEmail_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckExistingEmailRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).CheckExistingEmail(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_CheckExistingEmail_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).CheckExistingEmail(ctx, req.(*CheckExistingEmailRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_CheckExistingPhone_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckExistingPhoneRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).CheckExistingPhone(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_CheckExistingPhone_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).CheckExistingPhone(ctx, req.(*CheckExistingPhoneRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_CheckValidUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CheckValidUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).CheckValidUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_CheckValidUser_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).CheckValidUser(ctx, req.(*CheckValidUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_GetListAccountDisplayInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetListAccountDisplayInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).GetListAccountDisplayInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_GetListAccountDisplayInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).GetListAccountDisplayInfo(ctx, req.(*GetListAccountDisplayInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserService_GetAccountInfo_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetAccountInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).GetAccountInfo(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: UserService_GetAccountInfo_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserServiceServer).GetAccountInfo(ctx, req.(*GetAccountInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// UserService_ServiceDesc is the grpc.ServiceDesc for UserService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UserService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "user.UserService",
	HandlerType: (*UserServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Login",
			Handler:    _UserService_Login_Handler,
		},
		{
			MethodName: "Signup",
			Handler:    _UserService_Signup_Handler,
		},
		{
			MethodName: "CheckExistingUsername",
			Handler:    _UserService_CheckExistingUsername_Handler,
		},
		{
			MethodName: "CheckExistingEmail",
			Handler:    _UserService_CheckExistingEmail_Handler,
		},
		{
			MethodName: "CheckExistingPhone",
			Handler:    _UserService_CheckExistingPhone_Handler,
		},
		{
			MethodName: "CheckValidUser",
			Handler:    _UserService_CheckValidUser_Handler,
		},
		{
			MethodName: "GetListAccountDisplayInfo",
			Handler:    _UserService_GetListAccountDisplayInfo_Handler,
		},
		{
			MethodName: "GetAccountInfo",
			Handler:    _UserService_GetAccountInfo_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/user_service/user_service.proto",
}
