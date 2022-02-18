// Code generated by protoc-gen-go. DO NOT EDIT.
// source: cosmos/tx.proto

package cosmos

import (
	context "context"
	fmt "fmt"
	proto1 "github.com/KiraCore/interx/proto"
	proto "github.com/golang/protobuf/proto"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

type PostTransactionResult struct {
	Code                 string   `protobuf:"bytes,1,opt,name=code,proto3" json:"code,omitempty"`
	Data                 string   `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
	Log                  string   `protobuf:"bytes,3,opt,name=log,proto3" json:"log,omitempty"`
	Codespace            string   `protobuf:"bytes,4,opt,name=codespace,proto3" json:"codespace,omitempty"`
	Hash                 string   `protobuf:"bytes,5,opt,name=hash,proto3" json:"hash,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PostTransactionResult) Reset()         { *m = PostTransactionResult{} }
func (m *PostTransactionResult) String() string { return proto.CompactTextString(m) }
func (*PostTransactionResult) ProtoMessage()    {}
func (*PostTransactionResult) Descriptor() ([]byte, []int) {
	return fileDescriptor_ed5f703ac9d782b4, []int{0}
}

func (m *PostTransactionResult) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PostTransactionResult.Unmarshal(m, b)
}
func (m *PostTransactionResult) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PostTransactionResult.Marshal(b, m, deterministic)
}
func (m *PostTransactionResult) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PostTransactionResult.Merge(m, src)
}
func (m *PostTransactionResult) XXX_Size() int {
	return xxx_messageInfo_PostTransactionResult.Size(m)
}
func (m *PostTransactionResult) XXX_DiscardUnknown() {
	xxx_messageInfo_PostTransactionResult.DiscardUnknown(m)
}

var xxx_messageInfo_PostTransactionResult proto.InternalMessageInfo

func (m *PostTransactionResult) GetCode() string {
	if m != nil {
		return m.Code
	}
	return ""
}

func (m *PostTransactionResult) GetData() string {
	if m != nil {
		return m.Data
	}
	return ""
}

func (m *PostTransactionResult) GetLog() string {
	if m != nil {
		return m.Log
	}
	return ""
}

func (m *PostTransactionResult) GetCodespace() string {
	if m != nil {
		return m.Codespace
	}
	return ""
}

func (m *PostTransactionResult) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

// PostTransactionRequest is the request type for the tx/PostTransaction RPC method.
type PostTransactionRequest struct {
	// transaction hash.
	Tx                   []byte   `protobuf:"bytes,1,opt,name=tx,proto3" json:"tx,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *PostTransactionRequest) Reset()         { *m = PostTransactionRequest{} }
func (m *PostTransactionRequest) String() string { return proto.CompactTextString(m) }
func (*PostTransactionRequest) ProtoMessage()    {}
func (*PostTransactionRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ed5f703ac9d782b4, []int{1}
}

func (m *PostTransactionRequest) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PostTransactionRequest.Unmarshal(m, b)
}
func (m *PostTransactionRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PostTransactionRequest.Marshal(b, m, deterministic)
}
func (m *PostTransactionRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PostTransactionRequest.Merge(m, src)
}
func (m *PostTransactionRequest) XXX_Size() int {
	return xxx_messageInfo_PostTransactionRequest.Size(m)
}
func (m *PostTransactionRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_PostTransactionRequest.DiscardUnknown(m)
}

var xxx_messageInfo_PostTransactionRequest proto.InternalMessageInfo

func (m *PostTransactionRequest) GetTx() []byte {
	if m != nil {
		return m.Tx
	}
	return nil
}

// PostTransactionResponse is the response type for the tx/PostTransaction RPC method.
type PostTransactionResponse struct {
	ChainId              string                 `protobuf:"bytes,1,opt,name=chain_id,json=chainId,proto3" json:"chain_id,omitempty"`
	Block                uint64                 `protobuf:"varint,2,opt,name=block,proto3" json:"block,omitempty"`
	BlockTime            string                 `protobuf:"bytes,3,opt,name=block_time,json=blockTime,proto3" json:"block_time,omitempty"`
	Timestamp            uint64                 `protobuf:"varint,4,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Response             *PostTransactionResult `protobuf:"bytes,5,opt,name=response,proto3" json:"response,omitempty"`
	Error                *proto1.Error          `protobuf:"bytes,6,opt,name=error,proto3" json:"error,omitempty"`
	Signature            string                 `protobuf:"bytes,7,opt,name=signature,proto3" json:"signature,omitempty"`
	Hash                 string                 `protobuf:"bytes,8,opt,name=hash,proto3" json:"hash,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *PostTransactionResponse) Reset()         { *m = PostTransactionResponse{} }
func (m *PostTransactionResponse) String() string { return proto.CompactTextString(m) }
func (*PostTransactionResponse) ProtoMessage()    {}
func (*PostTransactionResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ed5f703ac9d782b4, []int{2}
}

func (m *PostTransactionResponse) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_PostTransactionResponse.Unmarshal(m, b)
}
func (m *PostTransactionResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_PostTransactionResponse.Marshal(b, m, deterministic)
}
func (m *PostTransactionResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_PostTransactionResponse.Merge(m, src)
}
func (m *PostTransactionResponse) XXX_Size() int {
	return xxx_messageInfo_PostTransactionResponse.Size(m)
}
func (m *PostTransactionResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_PostTransactionResponse.DiscardUnknown(m)
}

var xxx_messageInfo_PostTransactionResponse proto.InternalMessageInfo

func (m *PostTransactionResponse) GetChainId() string {
	if m != nil {
		return m.ChainId
	}
	return ""
}

func (m *PostTransactionResponse) GetBlock() uint64 {
	if m != nil {
		return m.Block
	}
	return 0
}

func (m *PostTransactionResponse) GetBlockTime() string {
	if m != nil {
		return m.BlockTime
	}
	return ""
}

func (m *PostTransactionResponse) GetTimestamp() uint64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *PostTransactionResponse) GetResponse() *PostTransactionResult {
	if m != nil {
		return m.Response
	}
	return nil
}

func (m *PostTransactionResponse) GetError() *proto1.Error {
	if m != nil {
		return m.Error
	}
	return nil
}

func (m *PostTransactionResponse) GetSignature() string {
	if m != nil {
		return m.Signature
	}
	return ""
}

func (m *PostTransactionResponse) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

func init() {
	proto.RegisterType((*PostTransactionResult)(nil), "cosmos.tx.PostTransactionResult")
	proto.RegisterType((*PostTransactionRequest)(nil), "cosmos.tx.PostTransactionRequest")
	proto.RegisterType((*PostTransactionResponse)(nil), "cosmos.tx.PostTransactionResponse")
}

func init() {
	proto.RegisterFile("cosmos/tx.proto", fileDescriptor_ed5f703ac9d782b4)
}

var fileDescriptor_ed5f703ac9d782b4 = []byte{
	// 485 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x93, 0xdd, 0x6e, 0xd3, 0x30,
	0x14, 0xc7, 0x95, 0xb4, 0xdd, 0x5a, 0x77, 0xa2, 0x9b, 0xc5, 0x47, 0xa8, 0x86, 0xd4, 0xe5, 0x6a,
	0x80, 0x9a, 0x40, 0xb8, 0x43, 0x08, 0x69, 0x54, 0x5c, 0x20, 0x6e, 0xa6, 0x68, 0x57, 0xdc, 0x4c,
	0x6e, 0x6a, 0x25, 0x16, 0x8d, 0x6d, 0xec, 0x53, 0xc8, 0x6e, 0xe1, 0x0d, 0xc6, 0x1d, 0x12, 0x8f,
	0xc0, 0xd3, 0xf0, 0x0a, 0x3c, 0x08, 0xf2, 0x71, 0x95, 0x4e, 0xfb, 0xe8, 0x55, 0x8e, 0x7f, 0xfe,
	0xdb, 0xe7, 0x7f, 0x8e, 0x4f, 0xc8, 0xa8, 0x50, 0xb6, 0x56, 0x36, 0x85, 0x26, 0xd1, 0x46, 0x81,
	0xa2, 0x03, 0x0f, 0x12, 0x68, 0xc6, 0x87, 0xa5, 0x52, 0xe5, 0x92, 0xa7, 0x4c, 0x8b, 0x94, 0x49,
	0xa9, 0x80, 0x81, 0x50, 0xd2, 0x7a, 0xe1, 0xd8, 0x7f, 0x8a, 0x69, 0xc9, 0xe5, 0x54, 0x69, 0x2e,
	0x99, 0x16, 0x5f, 0xb3, 0x54, 0x69, 0xd4, 0xdc, 0xa2, 0x27, 0x70, 0xa1, 0xb9, 0x8f, 0xe3, 0x1f,
	0x01, 0x79, 0x70, 0xaa, 0x2c, 0x9c, 0x19, 0x26, 0x2d, 0x2b, 0x9c, 0x2c, 0xe7, 0x76, 0xb5, 0x04,
	0x4a, 0x49, 0xb7, 0x50, 0x0b, 0x1e, 0x05, 0x93, 0xe0, 0x78, 0x90, 0x63, 0xec, 0xd8, 0x82, 0x01,
	0x8b, 0x42, 0xcf, 0x5c, 0x4c, 0xf7, 0x49, 0x67, 0xa9, 0xca, 0xa8, 0x83, 0xc8, 0x85, 0xf4, 0x90,
	0x0c, 0x9c, 0xda, 0x6a, 0x56, 0xf0, 0xa8, 0x8b, 0x7c, 0x03, 0xdc, 0x1d, 0x15, 0xb3, 0x55, 0xd4,
	0xf3, 0x77, 0xb8, 0x38, 0x3e, 0x26, 0x0f, 0x6f, 0x98, 0xf8, 0xb2, 0xe2, 0x16, 0xe8, 0x3d, 0x12,
	0x42, 0x83, 0x1e, 0xf6, 0xf2, 0x10, 0x9a, 0xf8, 0x57, 0x48, 0x1e, 0xdd, 0xf4, 0xab, 0x95, 0xb4,
	0x9c, 0x3e, 0x26, 0xfd, 0xa2, 0x62, 0x42, 0x9e, 0x8b, 0xc5, 0xda, 0xf5, 0x2e, 0xae, 0x3f, 0x2c,
	0xe8, 0x7d, 0xd2, 0x9b, 0x2f, 0x55, 0xf1, 0x19, 0x9d, 0x77, 0x73, 0xbf, 0xa0, 0x4f, 0x08, 0xc1,
	0xe0, 0x1c, 0x44, 0xcd, 0xd7, 0x15, 0x0c, 0x90, 0x9c, 0x89, 0x9a, 0xbb, 0x3a, 0xdc, 0x86, 0x05,
	0x56, 0x6b, 0xac, 0xa3, 0x9b, 0x6f, 0x00, 0x7d, 0x43, 0xfa, 0x66, 0x9d, 0x19, 0x6b, 0x19, 0x66,
	0x93, 0xa4, 0x7d, 0xb1, 0xe4, 0xd6, 0x9e, 0xe6, 0xed, 0x09, 0x7a, 0x44, 0x7a, 0xdc, 0x18, 0x65,
	0xa2, 0x1d, 0x3c, 0x3a, 0x4c, 0xf0, 0x4d, 0xde, 0x3b, 0x94, 0xfb, 0x1d, 0x97, 0xde, 0x8a, 0x52,
	0x32, 0x58, 0x19, 0x1e, 0xed, 0x7a, 0x73, 0x2d, 0x68, 0xdb, 0xd8, 0xdf, 0xb4, 0x31, 0xfb, 0x13,
	0x90, 0xe1, 0x95, 0xa4, 0xf4, 0x77, 0x40, 0x46, 0xd7, 0x8c, 0xd0, 0xa3, 0x6d, 0x26, 0xb1, 0xe7,
	0xe3, 0x78, 0x6b, 0x1d, 0xe8, 0x3f, 0x9e, 0x5d, 0x9e, 0x3c, 0x75, 0x4f, 0x43, 0xf7, 0x9d, 0x62,
	0x72, 0x45, 0x32, 0x3e, 0xb8, 0x4e, 0x92, 0xef, 0x7f, 0xff, 0xfd, 0x0c, 0x0f, 0xe2, 0x11, 0xce,
	0x6f, 0x3b, 0xe6, 0xf6, 0x9d, 0xbc, 0x3c, 0x79, 0x4b, 0x7b, 0x59, 0xe7, 0x65, 0xf2, 0xe2, 0x59,
	0x10, 0x98, 0x8c, 0xec, 0x95, 0xf9, 0xe9, 0x6c, 0x5a, 0x32, 0xe0, 0xdf, 0xd8, 0x05, 0x8d, 0x2b,
	0x00, 0x6d, 0x5f, 0xa7, 0x69, 0x29, 0xa0, 0x5a, 0xcd, 0x93, 0x42, 0xd5, 0xe9, 0x47, 0x61, 0xd8,
	0x4c, 0x19, 0x9e, 0x0a, 0x09, 0xdc, 0x34, 0x9f, 0x9e, 0xdf, 0xbd, 0x97, 0xe2, 0x84, 0xbb, 0xbf,
	0x62, 0x9d, 0x73, 0xbe, 0x83, 0xe4, 0xd5, 0xff, 0x00, 0x00, 0x00, 0xff, 0xff, 0x63, 0xc5, 0x2c,
	0x23, 0x6b, 0x03, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConnInterface

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion6

// TransactionClient is the client API for Transaction service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type TransactionClient interface {
	PostTransaction(ctx context.Context, in *PostTransactionRequest, opts ...grpc.CallOption) (*PostTransactionResponse, error)
}

type transactionClient struct {
	cc grpc.ClientConnInterface
}

func NewTransactionClient(cc grpc.ClientConnInterface) TransactionClient {
	return &transactionClient{cc}
}

func (c *transactionClient) PostTransaction(ctx context.Context, in *PostTransactionRequest, opts ...grpc.CallOption) (*PostTransactionResponse, error) {
	out := new(PostTransactionResponse)
	err := c.cc.Invoke(ctx, "/cosmos.tx.Transaction/PostTransaction", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TransactionServer is the server API for Transaction service.
type TransactionServer interface {
	PostTransaction(context.Context, *PostTransactionRequest) (*PostTransactionResponse, error)
}

// UnimplementedTransactionServer can be embedded to have forward compatible implementations.
type UnimplementedTransactionServer struct {
}

func (*UnimplementedTransactionServer) PostTransaction(ctx context.Context, req *PostTransactionRequest) (*PostTransactionResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method PostTransaction not implemented")
}

func RegisterTransactionServer(s *grpc.Server, srv TransactionServer) {
	s.RegisterService(&_Transaction_serviceDesc, srv)
}

func _Transaction_PostTransaction_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(PostTransactionRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TransactionServer).PostTransaction(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/cosmos.tx.Transaction/PostTransaction",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TransactionServer).PostTransaction(ctx, req.(*PostTransactionRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Transaction_serviceDesc = grpc.ServiceDesc{
	ServiceName: "cosmos.tx.Transaction",
	HandlerType: (*TransactionServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "PostTransaction",
			Handler:    _Transaction_PostTransaction_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "cosmos/tx.proto",
}