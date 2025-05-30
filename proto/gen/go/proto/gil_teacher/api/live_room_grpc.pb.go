// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.5.1
// - protoc             v5.29.3
// source: proto/gil_teacher/api/live_room.proto

package api

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
	LiveRoom_Create_FullMethodName = "/gil_teacher.api.user.LiveRoom/Create"
	LiveRoom_Info_FullMethodName   = "/gil_teacher.api.user.LiveRoom/Info"
	LiveRoom_List_FullMethodName   = "/gil_teacher.api.user.LiveRoom/List"
	LiveRoom_Edit_FullMethodName   = "/gil_teacher.api.user.LiveRoom/Edit"
	LiveRoom_Update_FullMethodName = "/gil_teacher.api.user.LiveRoom/Update"
)

// LiveRoomClient is the client API for LiveRoom service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type LiveRoomClient interface {
	// Create a new live room
	Create(ctx context.Context, in *LiveRoomCreateRequest, opts ...grpc.CallOption) (*LiveRoomCreateResponse, error)
	// Get live room info
	Info(ctx context.Context, in *LiveRoomInfoRequest, opts ...grpc.CallOption) (*LiveRoomInfoResponse, error)
	// Get live room list
	List(ctx context.Context, in *LiveRoomListRequest, opts ...grpc.CallOption) (*LiveRoomListResponse, error)
	// Edit live room
	Edit(ctx context.Context, in *LiveRoomEditRequest, opts ...grpc.CallOption) (*LiveRoomEditResponse, error)
	// Update live room
	Update(ctx context.Context, in *LiveRoomUpdateRequest, opts ...grpc.CallOption) (*LiveRoomUpdateResponse, error)
}

type liveRoomClient struct {
	cc grpc.ClientConnInterface
}

func NewLiveRoomClient(cc grpc.ClientConnInterface) LiveRoomClient {
	return &liveRoomClient{cc}
}

func (c *liveRoomClient) Create(ctx context.Context, in *LiveRoomCreateRequest, opts ...grpc.CallOption) (*LiveRoomCreateResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LiveRoomCreateResponse)
	err := c.cc.Invoke(ctx, LiveRoom_Create_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *liveRoomClient) Info(ctx context.Context, in *LiveRoomInfoRequest, opts ...grpc.CallOption) (*LiveRoomInfoResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LiveRoomInfoResponse)
	err := c.cc.Invoke(ctx, LiveRoom_Info_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *liveRoomClient) List(ctx context.Context, in *LiveRoomListRequest, opts ...grpc.CallOption) (*LiveRoomListResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LiveRoomListResponse)
	err := c.cc.Invoke(ctx, LiveRoom_List_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *liveRoomClient) Edit(ctx context.Context, in *LiveRoomEditRequest, opts ...grpc.CallOption) (*LiveRoomEditResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LiveRoomEditResponse)
	err := c.cc.Invoke(ctx, LiveRoom_Edit_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *liveRoomClient) Update(ctx context.Context, in *LiveRoomUpdateRequest, opts ...grpc.CallOption) (*LiveRoomUpdateResponse, error) {
	cOpts := append([]grpc.CallOption{grpc.StaticMethod()}, opts...)
	out := new(LiveRoomUpdateResponse)
	err := c.cc.Invoke(ctx, LiveRoom_Update_FullMethodName, in, out, cOpts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// LiveRoomServer is the server API for LiveRoom service.
// All implementations must embed UnimplementedLiveRoomServer
// for forward compatibility.
type LiveRoomServer interface {
	// Create a new live room
	Create(context.Context, *LiveRoomCreateRequest) (*LiveRoomCreateResponse, error)
	// Get live room info
	Info(context.Context, *LiveRoomInfoRequest) (*LiveRoomInfoResponse, error)
	// Get live room list
	List(context.Context, *LiveRoomListRequest) (*LiveRoomListResponse, error)
	// Edit live room
	Edit(context.Context, *LiveRoomEditRequest) (*LiveRoomEditResponse, error)
	// Update live room
	Update(context.Context, *LiveRoomUpdateRequest) (*LiveRoomUpdateResponse, error)
	mustEmbedUnimplementedLiveRoomServer()
}

// UnimplementedLiveRoomServer must be embedded to have
// forward compatible implementations.
//
// NOTE: this should be embedded by value instead of pointer to avoid a nil
// pointer dereference when methods are called.
type UnimplementedLiveRoomServer struct{}

func (UnimplementedLiveRoomServer) Create(context.Context, *LiveRoomCreateRequest) (*LiveRoomCreateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}
func (UnimplementedLiveRoomServer) Info(context.Context, *LiveRoomInfoRequest) (*LiveRoomInfoResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Info not implemented")
}
func (UnimplementedLiveRoomServer) List(context.Context, *LiveRoomListRequest) (*LiveRoomListResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method List not implemented")
}
func (UnimplementedLiveRoomServer) Edit(context.Context, *LiveRoomEditRequest) (*LiveRoomEditResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Edit not implemented")
}
func (UnimplementedLiveRoomServer) Update(context.Context, *LiveRoomUpdateRequest) (*LiveRoomUpdateResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}
func (UnimplementedLiveRoomServer) mustEmbedUnimplementedLiveRoomServer() {}
func (UnimplementedLiveRoomServer) testEmbeddedByValue()                  {}

// UnsafeLiveRoomServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to LiveRoomServer will
// result in compilation errors.
type UnsafeLiveRoomServer interface {
	mustEmbedUnimplementedLiveRoomServer()
}

func RegisterLiveRoomServer(s grpc.ServiceRegistrar, srv LiveRoomServer) {
	// If the following call pancis, it indicates UnimplementedLiveRoomServer was
	// embedded by pointer and is nil.  This will cause panics if an
	// unimplemented method is ever invoked, so we test this at initialization
	// time to prevent it from happening at runtime later due to I/O.
	if t, ok := srv.(interface{ testEmbeddedByValue() }); ok {
		t.testEmbeddedByValue()
	}
	s.RegisterService(&LiveRoom_ServiceDesc, srv)
}

func _LiveRoom_Create_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LiveRoomCreateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LiveRoomServer).Create(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LiveRoom_Create_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LiveRoomServer).Create(ctx, req.(*LiveRoomCreateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LiveRoom_Info_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LiveRoomInfoRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LiveRoomServer).Info(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LiveRoom_Info_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LiveRoomServer).Info(ctx, req.(*LiveRoomInfoRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LiveRoom_List_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LiveRoomListRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LiveRoomServer).List(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LiveRoom_List_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LiveRoomServer).List(ctx, req.(*LiveRoomListRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LiveRoom_Edit_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LiveRoomEditRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LiveRoomServer).Edit(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LiveRoom_Edit_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LiveRoomServer).Edit(ctx, req.(*LiveRoomEditRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _LiveRoom_Update_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(LiveRoomUpdateRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LiveRoomServer).Update(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: LiveRoom_Update_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LiveRoomServer).Update(ctx, req.(*LiveRoomUpdateRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// LiveRoom_ServiceDesc is the grpc.ServiceDesc for LiveRoom service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var LiveRoom_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "gil_teacher.api.user.LiveRoom",
	HandlerType: (*LiveRoomServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Create",
			Handler:    _LiveRoom_Create_Handler,
		},
		{
			MethodName: "Info",
			Handler:    _LiveRoom_Info_Handler,
		},
		{
			MethodName: "List",
			Handler:    _LiveRoom_List_Handler,
		},
		{
			MethodName: "Edit",
			Handler:    _LiveRoom_Edit_Handler,
		},
		{
			MethodName: "Update",
			Handler:    _LiveRoom_Update_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "proto/gil_teacher/api/live_room.proto",
}
