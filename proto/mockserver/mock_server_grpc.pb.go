// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v4.25.3
// source: mockserver/mock_server.proto

package mockserver

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

// MockServerClient is the client API for MockServer service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type MockServerClient interface {
	SetMockUrl(ctx context.Context, in *SetMockUrlRequest, opts ...grpc.CallOption) (*SetMockUrlResponse, error)
	GetMockResponse(ctx context.Context, in *MockRequest, opts ...grpc.CallOption) (*MockResponse, error)
}

type mockServerClient struct {
	cc grpc.ClientConnInterface
}

func NewMockServerClient(cc grpc.ClientConnInterface) MockServerClient {
	return &mockServerClient{cc}
}

func (c *mockServerClient) SetMockUrl(ctx context.Context, in *SetMockUrlRequest, opts ...grpc.CallOption) (*SetMockUrlResponse, error) {
	out := new(SetMockUrlResponse)
	err := c.cc.Invoke(ctx, "/mockserver.MockServer/SetMockUrl", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *mockServerClient) GetMockResponse(ctx context.Context, in *MockRequest, opts ...grpc.CallOption) (*MockResponse, error) {
	out := new(MockResponse)
	err := c.cc.Invoke(ctx, "/mockserver.MockServer/GetMockResponse", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MockServerServer is the server API for MockServer service.
// All implementations must embed UnimplementedMockServerServer
// for forward compatibility
type MockServerServer interface {
	SetMockUrl(context.Context, *SetMockUrlRequest) (*SetMockUrlResponse, error)
	GetMockResponse(context.Context, *MockRequest) (*MockResponse, error)
	mustEmbedUnimplementedMockServerServer()
}

// UnimplementedMockServerServer must be embedded to have forward compatible implementations.
type UnimplementedMockServerServer struct {
}

func (UnimplementedMockServerServer) SetMockUrl(context.Context, *SetMockUrlRequest) (*SetMockUrlResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SetMockUrl not implemented")
}
func (UnimplementedMockServerServer) GetMockResponse(context.Context, *MockRequest) (*MockResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetMockResponse not implemented")
}
func (UnimplementedMockServerServer) mustEmbedUnimplementedMockServerServer() {}

// UnsafeMockServerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to MockServerServer will
// result in compilation errors.
type UnsafeMockServerServer interface {
	mustEmbedUnimplementedMockServerServer()
}

func RegisterMockServerServer(s grpc.ServiceRegistrar, srv MockServerServer) {
	s.RegisterService(&MockServer_ServiceDesc, srv)
}

func _MockServer_SetMockUrl_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SetMockUrlRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MockServerServer).SetMockUrl(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mockserver.MockServer/SetMockUrl",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MockServerServer).SetMockUrl(ctx, req.(*SetMockUrlRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _MockServer_GetMockResponse_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MockRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MockServerServer).GetMockResponse(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/mockserver.MockServer/GetMockResponse",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MockServerServer).GetMockResponse(ctx, req.(*MockRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// MockServer_ServiceDesc is the grpc.ServiceDesc for MockServer service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var MockServer_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "mockserver.MockServer",
	HandlerType: (*MockServerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SetMockUrl",
			Handler:    _MockServer_SetMockUrl_Handler,
		},
		{
			MethodName: "GetMockResponse",
			Handler:    _MockServer_GetMockResponse_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "mockserver/mock_server.proto",
}
