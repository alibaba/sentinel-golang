// Code generated by protoc-gen-go. DO NOT EDIT.
// source: adapter/go_micro/proto/test.proto

package proto

import (
	fmt "fmt"
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

type Request struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Request) Reset()         { *m = Request{} }
func (m *Request) String() string { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()    {}
func (*Request) Descriptor() ([]byte, []int) {
	return fileDescriptor_5cc5df5bdcf2b490, []int{0}
}

func (m *Request) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Request.Unmarshal(m, b)
}
func (m *Request) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Request.Marshal(b, m, deterministic)
}
func (m *Request) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Request.Merge(m, src)
}
func (m *Request) XXX_Size() int {
	return xxx_messageInfo_Request.Size(m)
}
func (m *Request) XXX_DiscardUnknown() {
	xxx_messageInfo_Request.DiscardUnknown(m)
}

var xxx_messageInfo_Request proto.InternalMessageInfo

type Response struct {
	Result               string   `protobuf:"bytes,1,opt,name=result,proto3" json:"result,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Response) Reset()         { *m = Response{} }
func (m *Response) String() string { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()    {}
func (*Response) Descriptor() ([]byte, []int) {
	return fileDescriptor_5cc5df5bdcf2b490, []int{1}
}

func (m *Response) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Response.Unmarshal(m, b)
}
func (m *Response) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Response.Marshal(b, m, deterministic)
}
func (m *Response) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Response.Merge(m, src)
}
func (m *Response) XXX_Size() int {
	return xxx_messageInfo_Response.Size(m)
}
func (m *Response) XXX_DiscardUnknown() {
	xxx_messageInfo_Response.DiscardUnknown(m)
}

var xxx_messageInfo_Response proto.InternalMessageInfo

func (m *Response) GetResult() string {
	if m != nil {
		return m.Result
	}
	return ""
}

func init() {
	proto.RegisterType((*Request)(nil), "proto.Request")
	proto.RegisterType((*Response)(nil), "proto.Response")
}

func init() { proto.RegisterFile("adapter/go_micro/proto/test.proto", fileDescriptor_5cc5df5bdcf2b490) }

var fileDescriptor_5cc5df5bdcf2b490 = []byte{
	// 146 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x4c, 0x4c, 0x49, 0x2c,
	0x28, 0x49, 0x2d, 0xd2, 0x4f, 0xcf, 0x8f, 0xcf, 0xcd, 0x4c, 0x2e, 0xca, 0xd7, 0x2f, 0x28, 0xca,
	0x2f, 0xc9, 0xd7, 0x2f, 0x49, 0x2d, 0x2e, 0xd1, 0x03, 0x33, 0x85, 0x58, 0xc1, 0x94, 0x12, 0x27,
	0x17, 0x7b, 0x50, 0x6a, 0x61, 0x69, 0x6a, 0x71, 0x89, 0x92, 0x12, 0x17, 0x47, 0x50, 0x6a, 0x71,
	0x41, 0x7e, 0x5e, 0x71, 0xaa, 0x90, 0x18, 0x17, 0x5b, 0x51, 0x6a, 0x71, 0x69, 0x4e, 0x89, 0x04,
	0xa3, 0x02, 0xa3, 0x06, 0x67, 0x10, 0x94, 0x67, 0x64, 0xc8, 0xc5, 0x12, 0x92, 0x5a, 0x5c, 0x22,
	0xa4, 0xc9, 0xc5, 0x12, 0x90, 0x99, 0x97, 0x2e, 0xc4, 0x07, 0x31, 0x4d, 0x0f, 0x6a, 0x86, 0x14,
	0x3f, 0x9c, 0x0f, 0x31, 0x48, 0x89, 0xc1, 0x49, 0x2e, 0x4a, 0x06, 0xbb, 0x6b, 0xac, 0xc1, 0x64,
	0x12, 0x1b, 0x98, 0x32, 0x06, 0x04, 0x00, 0x00, 0xff, 0xff, 0x34, 0x4d, 0x11, 0xfd, 0xb4, 0x00,
	0x00, 0x00,
}
