// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.6.1
// source: userbalance.proto

package api

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

// UserBalanceClient is the client API for UserBalance service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type UserBalanceClient interface {
	GetBalance(ctx context.Context, in *User, opts ...grpc.CallOption) (*User, error)
	ReplenishmentBalance(ctx context.Context, in *Replenishment, opts ...grpc.CallOption) (*Response, error)
	Transfer(ctx context.Context, in *Money, opts ...grpc.CallOption) (*Response, error)
	GetHistory(ctx context.Context, in *RequestHistory, opts ...grpc.CallOption) (*Histories, error)
	Reservation(ctx context.Context, in *Transaction, opts ...grpc.CallOption) (*Response, error)
	Confirmation(ctx context.Context, in *Transaction, opts ...grpc.CallOption) (*Response, error)
	CancelReservation(ctx context.Context, in *Transaction, opts ...grpc.CallOption) (*Response, error)
}

type userBalanceClient struct {
	cc grpc.ClientConnInterface
}

func NewUserBalanceClient(cc grpc.ClientConnInterface) UserBalanceClient {
	return &userBalanceClient{cc}
}

func (c *userBalanceClient) GetBalance(ctx context.Context, in *User, opts ...grpc.CallOption) (*User, error) {
	out := new(User)
	err := c.cc.Invoke(ctx, "/api.UserBalance/GetBalance", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userBalanceClient) ReplenishmentBalance(ctx context.Context, in *Replenishment, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/api.UserBalance/ReplenishmentBalance", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userBalanceClient) Transfer(ctx context.Context, in *Money, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/api.UserBalance/Transfer", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userBalanceClient) GetHistory(ctx context.Context, in *RequestHistory, opts ...grpc.CallOption) (*Histories, error) {
	out := new(Histories)
	err := c.cc.Invoke(ctx, "/api.UserBalance/GetHistory", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userBalanceClient) Reservation(ctx context.Context, in *Transaction, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/api.UserBalance/Reservation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userBalanceClient) Confirmation(ctx context.Context, in *Transaction, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/api.UserBalance/Confirmation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userBalanceClient) CancelReservation(ctx context.Context, in *Transaction, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := c.cc.Invoke(ctx, "/api.UserBalance/CancelReservation", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// UserBalanceServer is the server API for UserBalance service.
// All implementations should embed UnimplementedUserBalanceServer
// for forward compatibility
type UserBalanceServer interface {
	GetBalance(context.Context, *User) (*User, error)
	ReplenishmentBalance(context.Context, *Replenishment) (*Response, error)
	Transfer(context.Context, *Money) (*Response, error)
	GetHistory(context.Context, *RequestHistory) (*Histories, error)
	Reservation(context.Context, *Transaction) (*Response, error)
	Confirmation(context.Context, *Transaction) (*Response, error)
	CancelReservation(context.Context, *Transaction) (*Response, error)
}

// UnimplementedUserBalanceServer should be embedded to have forward compatible implementations.
type UnimplementedUserBalanceServer struct {
}

func (UnimplementedUserBalanceServer) GetBalance(context.Context, *User) (*User, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetBalance not implemented")
}
func (UnimplementedUserBalanceServer) ReplenishmentBalance(context.Context, *Replenishment) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReplenishmentBalance not implemented")
}
func (UnimplementedUserBalanceServer) Transfer(context.Context, *Money) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Transfer not implemented")
}
func (UnimplementedUserBalanceServer) GetHistory(context.Context, *RequestHistory) (*Histories, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetHistory not implemented")
}
func (UnimplementedUserBalanceServer) Reservation(context.Context, *Transaction) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Reservation not implemented")
}
func (UnimplementedUserBalanceServer) Confirmation(context.Context, *Transaction) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Confirmation not implemented")
}
func (UnimplementedUserBalanceServer) CancelReservation(context.Context, *Transaction) (*Response, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CancelReservation not implemented")
}

// UnsafeUserBalanceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to UserBalanceServer will
// result in compilation errors.
type UnsafeUserBalanceServer interface {
	mustEmbedUnimplementedUserBalanceServer()
}

func RegisterUserBalanceServer(s grpc.ServiceRegistrar, srv UserBalanceServer) {
	s.RegisterService(&UserBalance_ServiceDesc, srv)
}

func _UserBalance_GetBalance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(User)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserBalanceServer).GetBalance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.UserBalance/GetBalance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserBalanceServer).GetBalance(ctx, req.(*User))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserBalance_ReplenishmentBalance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Replenishment)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserBalanceServer).ReplenishmentBalance(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.UserBalance/ReplenishmentBalance",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserBalanceServer).ReplenishmentBalance(ctx, req.(*Replenishment))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserBalance_Transfer_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Money)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserBalanceServer).Transfer(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.UserBalance/Transfer",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserBalanceServer).Transfer(ctx, req.(*Money))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserBalance_GetHistory_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RequestHistory)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserBalanceServer).GetHistory(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.UserBalance/GetHistory",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserBalanceServer).GetHistory(ctx, req.(*RequestHistory))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserBalance_Reservation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Transaction)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserBalanceServer).Reservation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.UserBalance/Reservation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserBalanceServer).Reservation(ctx, req.(*Transaction))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserBalance_Confirmation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Transaction)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserBalanceServer).Confirmation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.UserBalance/Confirmation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserBalanceServer).Confirmation(ctx, req.(*Transaction))
	}
	return interceptor(ctx, in, info, handler)
}

func _UserBalance_CancelReservation_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Transaction)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserBalanceServer).CancelReservation(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/api.UserBalance/CancelReservation",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(UserBalanceServer).CancelReservation(ctx, req.(*Transaction))
	}
	return interceptor(ctx, in, info, handler)
}

// UserBalance_ServiceDesc is the grpc.ServiceDesc for UserBalance service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var UserBalance_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "api.UserBalance",
	HandlerType: (*UserBalanceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "GetBalance",
			Handler:    _UserBalance_GetBalance_Handler,
		},
		{
			MethodName: "ReplenishmentBalance",
			Handler:    _UserBalance_ReplenishmentBalance_Handler,
		},
		{
			MethodName: "Transfer",
			Handler:    _UserBalance_Transfer_Handler,
		},
		{
			MethodName: "GetHistory",
			Handler:    _UserBalance_GetHistory_Handler,
		},
		{
			MethodName: "Reservation",
			Handler:    _UserBalance_Reservation_Handler,
		},
		{
			MethodName: "Confirmation",
			Handler:    _UserBalance_Confirmation_Handler,
		},
		{
			MethodName: "CancelReservation",
			Handler:    _UserBalance_CancelReservation_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "userbalance.proto",
}