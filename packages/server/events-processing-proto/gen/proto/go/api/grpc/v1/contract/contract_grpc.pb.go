// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             (unknown)
// source: contract.proto

package contract_grpc_service

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// ContractGrpcServiceClient is the client API for ContractGrpcService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type ContractGrpcServiceClient interface {
	RolloutRenewalOpportunityOnExpiration(ctx context.Context, in *RolloutRenewalOpportunityOnExpirationGrpcRequest, opts ...grpc.CallOption) (*ContractIdGrpcResponse, error)
	RefreshContractStatus(ctx context.Context, in *RefreshContractStatusGrpcRequest, opts ...grpc.CallOption) (*ContractIdGrpcResponse, error)
	RefreshContractLtv(ctx context.Context, in *RefreshContractLtvGrpcRequest, opts ...grpc.CallOption) (*ContractIdGrpcResponse, error)
	SoftDeleteContract(ctx context.Context, in *SoftDeleteContractGrpcRequest, opts ...grpc.CallOption) (*emptypb.Empty, error)
}

type contractGrpcServiceClient struct {
	cc grpc.ClientConnInterface
}

func NewContractGrpcServiceClient(cc grpc.ClientConnInterface) ContractGrpcServiceClient {
	return &contractGrpcServiceClient{cc}
}

func (c *contractGrpcServiceClient) RolloutRenewalOpportunityOnExpiration(ctx context.Context, in *RolloutRenewalOpportunityOnExpirationGrpcRequest, opts ...grpc.CallOption) (*ContractIdGrpcResponse, error) {
	out := new(ContractIdGrpcResponse)
	err := c.cc.Invoke(ctx, "/ContractGrpcService/RolloutRenewalOpportunityOnExpiration", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *contractGrpcServiceClient) RefreshContractStatus(ctx context.Context, in *RefreshContractStatusGrpcRequest, opts ...grpc.CallOption) (*ContractIdGrpcResponse, error) {
	out := new(ContractIdGrpcResponse)
	err := c.cc.Invoke(ctx, "/ContractGrpcService/RefreshContractStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *contractGrpcServiceClient) RefreshContractLtv(ctx context.Context, in *RefreshContractLtvGrpcRequest, opts ...grpc.CallOption) (*ContractIdGrpcResponse, error) {
	out := new(ContractIdGrpcResponse)
	err := c.cc.Invoke(ctx, "/ContractGrpcService/RefreshContractLtv", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *contractGrpcServiceClient) SoftDeleteContract(ctx context.Context, in *SoftDeleteContractGrpcRequest, opts ...grpc.CallOption) (*emptypb.Empty, error) {
	out := new(emptypb.Empty)
	err := c.cc.Invoke(ctx, "/ContractGrpcService/SoftDeleteContract", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ContractGrpcServiceServer is the server API for ContractGrpcService service.
// All implementations should embed UnimplementedContractGrpcServiceServer
// for forward compatibility
type ContractGrpcServiceServer interface {
	RolloutRenewalOpportunityOnExpiration(context.Context, *RolloutRenewalOpportunityOnExpirationGrpcRequest) (*ContractIdGrpcResponse, error)
	RefreshContractStatus(context.Context, *RefreshContractStatusGrpcRequest) (*ContractIdGrpcResponse, error)
	RefreshContractLtv(context.Context, *RefreshContractLtvGrpcRequest) (*ContractIdGrpcResponse, error)
	SoftDeleteContract(context.Context, *SoftDeleteContractGrpcRequest) (*emptypb.Empty, error)
}

// UnimplementedContractGrpcServiceServer should be embedded to have forward compatible implementations.
type UnimplementedContractGrpcServiceServer struct {
}

func (UnimplementedContractGrpcServiceServer) RolloutRenewalOpportunityOnExpiration(context.Context, *RolloutRenewalOpportunityOnExpirationGrpcRequest) (*ContractIdGrpcResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RolloutRenewalOpportunityOnExpiration not implemented")
}
func (UnimplementedContractGrpcServiceServer) RefreshContractStatus(context.Context, *RefreshContractStatusGrpcRequest) (*ContractIdGrpcResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RefreshContractStatus not implemented")
}
func (UnimplementedContractGrpcServiceServer) RefreshContractLtv(context.Context, *RefreshContractLtvGrpcRequest) (*ContractIdGrpcResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method RefreshContractLtv not implemented")
}
func (UnimplementedContractGrpcServiceServer) SoftDeleteContract(context.Context, *SoftDeleteContractGrpcRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SoftDeleteContract not implemented")
}

// UnsafeContractGrpcServiceServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to ContractGrpcServiceServer will
// result in compilation errors.
type UnsafeContractGrpcServiceServer interface {
	mustEmbedUnimplementedContractGrpcServiceServer()
}

func RegisterContractGrpcServiceServer(s grpc.ServiceRegistrar, srv ContractGrpcServiceServer) {
	s.RegisterService(&ContractGrpcService_ServiceDesc, srv)
}

func _ContractGrpcService_RolloutRenewalOpportunityOnExpiration_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RolloutRenewalOpportunityOnExpirationGrpcRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContractGrpcServiceServer).RolloutRenewalOpportunityOnExpiration(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ContractGrpcService/RolloutRenewalOpportunityOnExpiration",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContractGrpcServiceServer).RolloutRenewalOpportunityOnExpiration(ctx, req.(*RolloutRenewalOpportunityOnExpirationGrpcRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ContractGrpcService_RefreshContractStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RefreshContractStatusGrpcRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContractGrpcServiceServer).RefreshContractStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ContractGrpcService/RefreshContractStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContractGrpcServiceServer).RefreshContractStatus(ctx, req.(*RefreshContractStatusGrpcRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ContractGrpcService_RefreshContractLtv_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RefreshContractLtvGrpcRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContractGrpcServiceServer).RefreshContractLtv(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ContractGrpcService/RefreshContractLtv",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContractGrpcServiceServer).RefreshContractLtv(ctx, req.(*RefreshContractLtvGrpcRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _ContractGrpcService_SoftDeleteContract_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(SoftDeleteContractGrpcRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContractGrpcServiceServer).SoftDeleteContract(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/ContractGrpcService/SoftDeleteContract",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContractGrpcServiceServer).SoftDeleteContract(ctx, req.(*SoftDeleteContractGrpcRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// ContractGrpcService_ServiceDesc is the grpc.ServiceDesc for ContractGrpcService service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var ContractGrpcService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "ContractGrpcService",
	HandlerType: (*ContractGrpcServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RolloutRenewalOpportunityOnExpiration",
			Handler:    _ContractGrpcService_RolloutRenewalOpportunityOnExpiration_Handler,
		},
		{
			MethodName: "RefreshContractStatus",
			Handler:    _ContractGrpcService_RefreshContractStatus_Handler,
		},
		{
			MethodName: "RefreshContractLtv",
			Handler:    _ContractGrpcService_RefreshContractLtv_Handler,
		},
		{
			MethodName: "SoftDeleteContract",
			Handler:    _ContractGrpcService_SoftDeleteContract_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "contract.proto",
}
