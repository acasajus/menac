// Code generated by protoc-gen-go.
// source: coord/proto/coord.proto
// DO NOT EDIT!

/*
Package coord is a generated protocol buffer package.

It is generated from these files:
	coord/proto/coord.proto

It has these top-level messages:
	PeerInfo
	PeerInfoList
	RaftStep
	ProcessRaftResponse
*/
package coord

import proto "github.com/golang/protobuf/proto"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal

type PeerInfo struct {
	Id      uint64 `protobuf:"varint,1,opt" json:"Id,omitempty"`
	Address string `protobuf:"bytes,2,opt" json:"Address,omitempty"`
}

func (m *PeerInfo) Reset()         { *m = PeerInfo{} }
func (m *PeerInfo) String() string { return proto.CompactTextString(m) }
func (*PeerInfo) ProtoMessage()    {}

type PeerInfoList struct {
	Peers []*PeerInfo `protobuf:"bytes,1,rep,name=peers" json:"peers,omitempty"`
}

func (m *PeerInfoList) Reset()         { *m = PeerInfoList{} }
func (m *PeerInfoList) String() string { return proto.CompactTextString(m) }
func (*PeerInfoList) ProtoMessage()    {}

func (m *PeerInfoList) GetPeers() []*PeerInfo {
	if m != nil {
		return m.Peers
	}
	return nil
}

type RaftStep struct {
	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (m *RaftStep) Reset()         { *m = RaftStep{} }
func (m *RaftStep) String() string { return proto.CompactTextString(m) }
func (*RaftStep) ProtoMessage()    {}

type ProcessRaftResponse struct {
}

func (m *ProcessRaftResponse) Reset()         { *m = ProcessRaftResponse{} }
func (m *ProcessRaftResponse) String() string { return proto.CompactTextString(m) }
func (*ProcessRaftResponse) ProtoMessage()    {}

func init() {
}

// Client API for Coordinate service

type CoordinateClient interface {
	// Process a raft step
	EmitRaftStep(ctx context.Context, opts ...grpc.CallOption) (Coordinate_EmitRaftStepClient, error)
	// Register into the cluster and get a list of peers
	Register(ctx context.Context, in *PeerInfo, opts ...grpc.CallOption) (*PeerInfoList, error)
}

type coordinateClient struct {
	cc *grpc.ClientConn
}

func NewCoordinateClient(cc *grpc.ClientConn) CoordinateClient {
	return &coordinateClient{cc}
}

func (c *coordinateClient) EmitRaftStep(ctx context.Context, opts ...grpc.CallOption) (Coordinate_EmitRaftStepClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Coordinate_serviceDesc.Streams[0], c.cc, "/coord.Coordinate/EmitRaftStep", opts...)
	if err != nil {
		return nil, err
	}
	x := &coordinateEmitRaftStepClient{stream}
	return x, nil
}

type Coordinate_EmitRaftStepClient interface {
	Send(*RaftStep) error
	Recv() (*RaftStep, error)
	grpc.ClientStream
}

type coordinateEmitRaftStepClient struct {
	grpc.ClientStream
}

func (x *coordinateEmitRaftStepClient) Send(m *RaftStep) error {
	return x.ClientStream.SendProto(m)
}

func (x *coordinateEmitRaftStepClient) Recv() (*RaftStep, error) {
	m := new(RaftStep)
	if err := x.ClientStream.RecvProto(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *coordinateClient) Register(ctx context.Context, in *PeerInfo, opts ...grpc.CallOption) (*PeerInfoList, error) {
	out := new(PeerInfoList)
	err := grpc.Invoke(ctx, "/coord.Coordinate/Register", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Coordinate service

type CoordinateServer interface {
	// Process a raft step
	EmitRaftStep(Coordinate_EmitRaftStepServer) error
	// Register into the cluster and get a list of peers
	Register(context.Context, *PeerInfo) (*PeerInfoList, error)
}

func RegisterCoordinateServer(s *grpc.Server, srv CoordinateServer) {
	s.RegisterService(&_Coordinate_serviceDesc, srv)
}

func _Coordinate_EmitRaftStep_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(CoordinateServer).EmitRaftStep(&coordinateEmitRaftStepServer{stream})
}

type Coordinate_EmitRaftStepServer interface {
	Send(*RaftStep) error
	Recv() (*RaftStep, error)
	grpc.ServerStream
}

type coordinateEmitRaftStepServer struct {
	grpc.ServerStream
}

func (x *coordinateEmitRaftStepServer) Send(m *RaftStep) error {
	return x.ServerStream.SendProto(m)
}

func (x *coordinateEmitRaftStepServer) Recv() (*RaftStep, error) {
	m := new(RaftStep)
	if err := x.ServerStream.RecvProto(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _Coordinate_Register_Handler(srv interface{}, ctx context.Context, buf []byte) (proto.Message, error) {
	in := new(PeerInfo)
	if err := proto.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(CoordinateServer).Register(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

var _Coordinate_serviceDesc = grpc.ServiceDesc{
	ServiceName: "coord.Coordinate",
	HandlerType: (*CoordinateServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Register",
			Handler:    _Coordinate_Register_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "EmitRaftStep",
			Handler:       _Coordinate_EmitRaftStep_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
}