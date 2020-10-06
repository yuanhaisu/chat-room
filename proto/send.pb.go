// Copyright 2015 gRPC authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.13.0
// source: send.proto

package proto

import (
	context "context"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

var File_send_proto protoreflect.FileDescriptor

var file_send_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x73, 0x65, 0x6e, 0x64, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x73, 0x65,
	0x6e, 0x64, 0x1a, 0x09, 0x72, 0x65, 0x71, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x32, 0x38, 0x0a,
	0x04, 0x53, 0x65, 0x6e, 0x64, 0x12, 0x30, 0x0a, 0x0e, 0x55, 0x73, 0x65, 0x72, 0x53, 0x65, 0x6e,
	0x64, 0x53, 0x74, 0x72, 0x65, 0x61, 0x6d, 0x12, 0x0c, 0x2e, 0x72, 0x65, 0x71, 0x2e, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x0c, 0x2e, 0x72, 0x65, 0x71, 0x2e, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x22, 0x00, 0x30, 0x01, 0x42, 0x11, 0x5a, 0x0f, 0x63, 0x68, 0x61, 0x74, 0x2d,
	0x72, 0x6f, 0x6f, 0x6d, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var file_send_proto_goTypes = []interface{}{
	(*Request)(nil), // 0: req.Request
}
var file_send_proto_depIdxs = []int32{
	0, // 0: send.Send.UserSendStream:input_type -> req.Request
	0, // 1: send.Send.UserSendStream:output_type -> req.Request
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_send_proto_init() }
func file_send_proto_init() {
	if File_send_proto != nil {
		return
	}
	file_req_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_send_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_send_proto_goTypes,
		DependencyIndexes: file_send_proto_depIdxs,
	}.Build()
	File_send_proto = out.File
	file_send_proto_rawDesc = nil
	file_send_proto_goTypes = nil
	file_send_proto_depIdxs = nil
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// SendClient is the client API for Send service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type SendClient interface {
	// Sends a greeting
	UserSendStream(ctx context.Context, in *Request, opts ...grpc.CallOption) (Send_UserSendStreamClient, error)
}

type sendClient struct {
	cc grpc.ClientConnInterface
}

func NewSendClient(cc grpc.ClientConnInterface) SendClient {
	return &sendClient{cc}
}

func (c *sendClient) UserSendStream(ctx context.Context, in *Request, opts ...grpc.CallOption) (Send_UserSendStreamClient, error) {
	stream, err := c.cc.NewStream(ctx, &_Send_serviceDesc.Streams[0], "/send.Send/UserSendStream", opts...)
	if err != nil {
		return nil, err
	}
	x := &sendUserSendStreamClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Send_UserSendStreamClient interface {
	Recv() (*Request, error)
	grpc.ClientStream
}

type sendUserSendStreamClient struct {
	grpc.ClientStream
}

func (x *sendUserSendStreamClient) Recv() (*Request, error) {
	m := new(Request)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// SendServer is the server API for Send service.
type SendServer interface {
	// Sends a greeting
	UserSendStream(*Request, Send_UserSendStreamServer) error
}

// UnimplementedSendServer can be embedded to have forward compatible implementations.
type UnimplementedSendServer struct {
}

func (*UnimplementedSendServer) UserSendStream(*Request, Send_UserSendStreamServer) error {
	return status.Errorf(codes.Unimplemented, "method UserSendStream not implemented")
}

func RegisterSendServer(s *grpc.Server, srv SendServer) {
	s.RegisterService(&_Send_serviceDesc, srv)
}

func _Send_UserSendStream_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Request)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SendServer).UserSendStream(m, &sendUserSendStreamServer{stream})
}

type Send_UserSendStreamServer interface {
	Send(*Request) error
	grpc.ServerStream
}

type sendUserSendStreamServer struct {
	grpc.ServerStream
}

func (x *sendUserSendStreamServer) Send(m *Request) error {
	return x.ServerStream.SendMsg(m)
}

var _Send_serviceDesc = grpc.ServiceDesc{
	ServiceName: "send.Send",
	HandlerType: (*SendServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "UserSendStream",
			Handler:       _Send_UserSendStream_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "send.proto",
}
