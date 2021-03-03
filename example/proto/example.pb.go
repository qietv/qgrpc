// Code generated by protoc-gen-go. DO NOT EDIT.
// source: example.proto

package proto

import (
	context "context"
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type TestReq struct {
	Ping                 string   `protobuf:"bytes,1,opt,name=ping,proto3" json:"ping,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TestReq) Reset()         { *m = TestReq{} }
func (m *TestReq) String() string { return proto.CompactTextString(m) }
func (*TestReq) ProtoMessage()    {}
func (*TestReq) Descriptor() ([]byte, []int) {
	return fileDescriptor_15a1dc8d40dadaa6, []int{0}
}

func (m *TestReq) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TestReq.Unmarshal(m, b)
}
func (m *TestReq) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TestReq.Marshal(b, m, deterministic)
}
func (m *TestReq) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TestReq.Merge(m, src)
}
func (m *TestReq) XXX_Size() int {
	return xxx_messageInfo_TestReq.Size(m)
}
func (m *TestReq) XXX_DiscardUnknown() {
	xxx_messageInfo_TestReq.DiscardUnknown(m)
}

var xxx_messageInfo_TestReq proto.InternalMessageInfo

func (m *TestReq) GetPing() string {
	if m != nil {
		return m.Ping
	}
	return ""
}

type TestResp struct {
	Pong                 string   `protobuf:"bytes,1,opt,name=pong,proto3" json:"pong,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TestResp) Reset()         { *m = TestResp{} }
func (m *TestResp) String() string { return proto.CompactTextString(m) }
func (*TestResp) ProtoMessage()    {}
func (*TestResp) Descriptor() ([]byte, []int) {
	return fileDescriptor_15a1dc8d40dadaa6, []int{1}
}

func (m *TestResp) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_TestResp.Unmarshal(m, b)
}
func (m *TestResp) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_TestResp.Marshal(b, m, deterministic)
}
func (m *TestResp) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TestResp.Merge(m, src)
}
func (m *TestResp) XXX_Size() int {
	return xxx_messageInfo_TestResp.Size(m)
}
func (m *TestResp) XXX_DiscardUnknown() {
	xxx_messageInfo_TestResp.DiscardUnknown(m)
}

var xxx_messageInfo_TestResp proto.InternalMessageInfo

func (m *TestResp) GetPong() string {
	if m != nil {
		return m.Pong
	}
	return ""
}

func init() {
	proto.RegisterType((*TestReq)(nil), "proto.TestReq")
	proto.RegisterType((*TestResp)(nil), "proto.TestResp")
}

func init() { proto.RegisterFile("example.proto", fileDescriptor_15a1dc8d40dadaa6) }

var fileDescriptor_15a1dc8d40dadaa6 = []byte{
	// 121 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x4d, 0xad, 0x48, 0xcc,
	0x2d, 0xc8, 0x49, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x05, 0x53, 0x4a, 0xb2, 0x5c,
	0xec, 0x21, 0xa9, 0xc5, 0x25, 0x41, 0xa9, 0x85, 0x42, 0x42, 0x5c, 0x2c, 0x05, 0x99, 0x79, 0xe9,
	0x12, 0x8c, 0x0a, 0x8c, 0x1a, 0x9c, 0x41, 0x60, 0xb6, 0x92, 0x1c, 0x17, 0x07, 0x44, 0xba, 0xb8,
	0x00, 0x2c, 0x9f, 0x8f, 0x24, 0x9f, 0x9f, 0x97, 0x6e, 0xa4, 0xcf, 0xc5, 0x02, 0x92, 0x17, 0x52,
	0xe7, 0x62, 0x29, 0x01, 0xd1, 0x7c, 0x10, 0xd3, 0xf5, 0xa0, 0x66, 0x4a, 0xf1, 0xa3, 0xf0, 0x8b,
	0x0b, 0x9c, 0xd8, 0xa3, 0x20, 0x16, 0x27, 0xb1, 0x81, 0x29, 0x63, 0x40, 0x00, 0x00, 0x00, 0xff,
	0xff, 0xa2, 0xd8, 0x5c, 0xc9, 0x97, 0x00, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// TestClient is the client API for Test service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TestClient interface {
	Test(ctx context.Context, in *TestReq, opts ...grpc.CallOption) (*TestResp, error)
}

type testClient struct {
	cc *grpc.ClientConn
}

func NewTestClient(cc *grpc.ClientConn) TestClient {
	return &testClient{cc}
}

func (c *testClient) Test(ctx context.Context, in *TestReq, opts ...grpc.CallOption) (*TestResp, error) {
	out := new(TestResp)
	err := c.cc.Invoke(ctx, "/proto.Test/test", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TestServer is the server API for Test service.
type TestServer interface {
	Test(context.Context, *TestReq) (*TestResp, error)
}

// UnimplementedTestServer can be embedded to have forward compatible implementations.
type UnimplementedTestServer struct {
}

func (*UnimplementedTestServer) Test(ctx context.Context, req *TestReq) (*TestResp, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Test not implemented")
}

func RegisterTestServer(s *grpc.Server, srv TestServer) {
	s.RegisterService(&_Test_serviceDesc, srv)
}

func _Test_Test_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TestReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TestServer).Test(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/proto.Test/Test",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TestServer).Test(ctx, req.(*TestReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _Test_serviceDesc = grpc.ServiceDesc{
	ServiceName: "proto.Test",
	HandlerType: (*TestServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "test",
			Handler:    _Test_Test_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "example.proto",
}