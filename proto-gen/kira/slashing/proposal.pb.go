// Code generated by protoc-gen-go. DO NOT EDIT.
// source: kira/slashing/proposal.proto

package slashing

import (
	fmt "fmt"
	_ "github.com/KiraCore/interx/proto-gen/kira/gov"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/golang/protobuf/proto"
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

type ProposalResetWholeValidatorRank struct {
	Proposer             []byte   `protobuf:"bytes,1,opt,name=proposer,proto3" json:"proposer,omitempty"`
	Description          string   `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ProposalResetWholeValidatorRank) Reset()         { *m = ProposalResetWholeValidatorRank{} }
func (m *ProposalResetWholeValidatorRank) String() string { return proto.CompactTextString(m) }
func (*ProposalResetWholeValidatorRank) ProtoMessage()    {}
func (*ProposalResetWholeValidatorRank) Descriptor() ([]byte, []int) {
	return fileDescriptor_c8be292de7dc6a45, []int{0}
}

func (m *ProposalResetWholeValidatorRank) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ProposalResetWholeValidatorRank.Unmarshal(m, b)
}
func (m *ProposalResetWholeValidatorRank) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ProposalResetWholeValidatorRank.Marshal(b, m, deterministic)
}
func (m *ProposalResetWholeValidatorRank) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ProposalResetWholeValidatorRank.Merge(m, src)
}
func (m *ProposalResetWholeValidatorRank) XXX_Size() int {
	return xxx_messageInfo_ProposalResetWholeValidatorRank.Size(m)
}
func (m *ProposalResetWholeValidatorRank) XXX_DiscardUnknown() {
	xxx_messageInfo_ProposalResetWholeValidatorRank.DiscardUnknown(m)
}

var xxx_messageInfo_ProposalResetWholeValidatorRank proto.InternalMessageInfo

func (m *ProposalResetWholeValidatorRank) GetProposer() []byte {
	if m != nil {
		return m.Proposer
	}
	return nil
}

func (m *ProposalResetWholeValidatorRank) GetDescription() string {
	if m != nil {
		return m.Description
	}
	return ""
}

func init() {
	proto.RegisterType((*ProposalResetWholeValidatorRank)(nil), "kira.slashing.ProposalResetWholeValidatorRank")
}

func init() {
	proto.RegisterFile("kira/slashing/proposal.proto", fileDescriptor_c8be292de7dc6a45)
}

var fileDescriptor_c8be292de7dc6a45 = []byte{
	// 235 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x90, 0xcf, 0x4a, 0xc3, 0x40,
	0x10, 0xc6, 0x89, 0x07, 0xd1, 0xa8, 0x97, 0x20, 0x58, 0x8a, 0xd0, 0xe0, 0xa9, 0x97, 0x64, 0xf0,
	0xcf, 0x0b, 0xb4, 0x1e, 0x45, 0x90, 0x1c, 0x14, 0xbc, 0x6d, 0x77, 0x87, 0xcd, 0x90, 0x74, 0x67,
	0x99, 0x59, 0x45, 0x5f, 0xc3, 0x07, 0xf4, 0x41, 0x3c, 0x49, 0xd3, 0x2a, 0xd1, 0xd3, 0x2e, 0xdf,
	0x6f, 0xbe, 0x1f, 0xc3, 0xe4, 0xe7, 0x1d, 0x89, 0x01, 0xed, 0x8d, 0xb6, 0x14, 0x3c, 0x44, 0xe1,
	0xc8, 0x6a, 0xfa, 0x3a, 0x0a, 0x27, 0x2e, 0x4e, 0x36, 0xb4, 0xfe, 0xa1, 0xd3, 0x53, 0xcf, 0x9e,
	0x07, 0x02, 0x9b, 0xdf, 0x76, 0x68, 0x7a, 0x36, 0x28, 0x3c, 0xbf, 0xfe, 0x6b, 0x5f, 0x7c, 0x64,
	0xf9, 0xec, 0x61, 0x17, 0x35, 0xa8, 0x98, 0x9e, 0x5a, 0xee, 0xf1, 0xd1, 0xf4, 0xe4, 0x4c, 0x62,
	0x69, 0x4c, 0xe8, 0x8a, 0xfb, 0xfc, 0x60, 0xdb, 0x42, 0x99, 0x64, 0x65, 0x36, 0x3f, 0x5e, 0x5e,
	0x7e, 0x7d, 0xce, 0x2a, 0x4f, 0xa9, 0x7d, 0x59, 0xd5, 0x96, 0xd7, 0x60, 0x59, 0xd7, 0xac, 0xbb,
	0xa7, 0x52, 0xd7, 0x41, 0x7a, 0x8f, 0xa8, 0xf5, 0xc2, 0xda, 0x85, 0x73, 0x82, 0xaa, 0xcd, 0xaf,
	0xa2, 0x28, 0xf3, 0x23, 0x87, 0x6a, 0x85, 0x62, 0x22, 0x0e, 0x93, 0xbd, 0x32, 0x9b, 0x1f, 0x36,
	0xe3, 0x68, 0x79, 0xf3, 0x7c, 0x35, 0x92, 0xdf, 0x91, 0x98, 0x5b, 0x16, 0x04, 0x0a, 0x09, 0xe5,
	0x0d, 0x86, 0xc5, 0x2b, 0x8f, 0x01, 0xfe, 0xdc, 0x65, 0xb5, 0x3f, 0x80, 0xeb, 0xef, 0x00, 0x00,
	0x00, 0xff, 0xff, 0x08, 0x21, 0x65, 0x3f, 0x2f, 0x01, 0x00, 0x00,
}