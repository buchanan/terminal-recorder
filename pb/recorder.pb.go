// Code generated by protoc-gen-go. DO NOT EDIT.
// source: recorder.proto

package recorder

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
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

type COMMAND struct {
	Line                 string               `protobuf:"bytes,1,opt,name=line,proto3" json:"line,omitempty"`
	Count                uint32               `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
	Timestamp            *timestamp.Timestamp `protobuf:"bytes,3,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Keystrokes           []*COMMAND_KEY       `protobuf:"bytes,4,rep,name=keystrokes,proto3" json:"keystrokes,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *COMMAND) Reset()         { *m = COMMAND{} }
func (m *COMMAND) String() string { return proto.CompactTextString(m) }
func (*COMMAND) ProtoMessage()    {}
func (*COMMAND) Descriptor() ([]byte, []int) {
	return fileDescriptor_b063ffe85a4e6395, []int{0}
}

func (m *COMMAND) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_COMMAND.Unmarshal(m, b)
}
func (m *COMMAND) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_COMMAND.Marshal(b, m, deterministic)
}
func (m *COMMAND) XXX_Merge(src proto.Message) {
	xxx_messageInfo_COMMAND.Merge(m, src)
}
func (m *COMMAND) XXX_Size() int {
	return xxx_messageInfo_COMMAND.Size(m)
}
func (m *COMMAND) XXX_DiscardUnknown() {
	xxx_messageInfo_COMMAND.DiscardUnknown(m)
}

var xxx_messageInfo_COMMAND proto.InternalMessageInfo

func (m *COMMAND) GetLine() string {
	if m != nil {
		return m.Line
	}
	return ""
}

func (m *COMMAND) GetCount() uint32 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *COMMAND) GetTimestamp() *timestamp.Timestamp {
	if m != nil {
		return m.Timestamp
	}
	return nil
}

func (m *COMMAND) GetKeystrokes() []*COMMAND_KEY {
	if m != nil {
		return m.Keystrokes
	}
	return nil
}

type COMMAND_KEY struct {
	Offset               float64  `protobuf:"fixed64,1,opt,name=offset,proto3" json:"offset,omitempty"`
	Tag                  string   `protobuf:"bytes,2,opt,name=tag,proto3" json:"tag,omitempty"`
	Key                  string   `protobuf:"bytes,3,opt,name=key,proto3" json:"key,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *COMMAND_KEY) Reset()         { *m = COMMAND_KEY{} }
func (m *COMMAND_KEY) String() string { return proto.CompactTextString(m) }
func (*COMMAND_KEY) ProtoMessage()    {}
func (*COMMAND_KEY) Descriptor() ([]byte, []int) {
	return fileDescriptor_b063ffe85a4e6395, []int{0, 0}
}

func (m *COMMAND_KEY) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_COMMAND_KEY.Unmarshal(m, b)
}
func (m *COMMAND_KEY) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_COMMAND_KEY.Marshal(b, m, deterministic)
}
func (m *COMMAND_KEY) XXX_Merge(src proto.Message) {
	xxx_messageInfo_COMMAND_KEY.Merge(m, src)
}
func (m *COMMAND_KEY) XXX_Size() int {
	return xxx_messageInfo_COMMAND_KEY.Size(m)
}
func (m *COMMAND_KEY) XXX_DiscardUnknown() {
	xxx_messageInfo_COMMAND_KEY.DiscardUnknown(m)
}

var xxx_messageInfo_COMMAND_KEY proto.InternalMessageInfo

func (m *COMMAND_KEY) GetOffset() float64 {
	if m != nil {
		return m.Offset
	}
	return 0
}

func (m *COMMAND_KEY) GetTag() string {
	if m != nil {
		return m.Tag
	}
	return ""
}

func (m *COMMAND_KEY) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

type HEADER struct {
	Version              uint32               `protobuf:"varint,1,opt,name=version,proto3" json:"version,omitempty"`
	Width                uint32               `protobuf:"varint,2,opt,name=width,proto3" json:"width,omitempty"`
	Height               uint32               `protobuf:"varint,3,opt,name=height,proto3" json:"height,omitempty"`
	Timestamp            *timestamp.Timestamp `protobuf:"bytes,4,opt,name=timestamp,proto3" json:"timestamp,omitempty"`
	Idle                 float64              `protobuf:"fixed64,5,opt,name=idle,proto3" json:"idle,omitempty"`
	Command              string               `protobuf:"bytes,6,opt,name=command,proto3" json:"command,omitempty"`
	Title                string               `protobuf:"bytes,7,opt,name=title,proto3" json:"title,omitempty"`
	Env                  []byte               `protobuf:"bytes,8,opt,name=env,proto3" json:"env,omitempty"`
	Theme                []byte               `protobuf:"bytes,9,opt,name=theme,proto3" json:"theme,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *HEADER) Reset()         { *m = HEADER{} }
func (m *HEADER) String() string { return proto.CompactTextString(m) }
func (*HEADER) ProtoMessage()    {}
func (*HEADER) Descriptor() ([]byte, []int) {
	return fileDescriptor_b063ffe85a4e6395, []int{1}
}

func (m *HEADER) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_HEADER.Unmarshal(m, b)
}
func (m *HEADER) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_HEADER.Marshal(b, m, deterministic)
}
func (m *HEADER) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HEADER.Merge(m, src)
}
func (m *HEADER) XXX_Size() int {
	return xxx_messageInfo_HEADER.Size(m)
}
func (m *HEADER) XXX_DiscardUnknown() {
	xxx_messageInfo_HEADER.DiscardUnknown(m)
}

var xxx_messageInfo_HEADER proto.InternalMessageInfo

func (m *HEADER) GetVersion() uint32 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *HEADER) GetWidth() uint32 {
	if m != nil {
		return m.Width
	}
	return 0
}

func (m *HEADER) GetHeight() uint32 {
	if m != nil {
		return m.Height
	}
	return 0
}

func (m *HEADER) GetTimestamp() *timestamp.Timestamp {
	if m != nil {
		return m.Timestamp
	}
	return nil
}

func (m *HEADER) GetIdle() float64 {
	if m != nil {
		return m.Idle
	}
	return 0
}

func (m *HEADER) GetCommand() string {
	if m != nil {
		return m.Command
	}
	return ""
}

func (m *HEADER) GetTitle() string {
	if m != nil {
		return m.Title
	}
	return ""
}

func (m *HEADER) GetEnv() []byte {
	if m != nil {
		return m.Env
	}
	return nil
}

func (m *HEADER) GetTheme() []byte {
	if m != nil {
		return m.Theme
	}
	return nil
}

func init() {
	proto.RegisterType((*COMMAND)(nil), "recorder.COMMAND")
	proto.RegisterType((*COMMAND_KEY)(nil), "recorder.COMMAND.KEY")
	proto.RegisterType((*HEADER)(nil), "recorder.HEADER")
}

func init() { proto.RegisterFile("recorder.proto", fileDescriptor_b063ffe85a4e6395) }

var fileDescriptor_b063ffe85a4e6395 = []byte{
	// 333 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x90, 0x41, 0x6b, 0xea, 0x40,
	0x10, 0xc7, 0xd9, 0x97, 0x18, 0xcd, 0xf8, 0x7c, 0x3c, 0x96, 0xb6, 0x2c, 0x5e, 0x1a, 0x3c, 0xe5,
	0x14, 0xc1, 0x52, 0xe8, 0x55, 0x6a, 0xa0, 0x20, 0xb6, 0xb0, 0xf4, 0xe2, 0x51, 0xcd, 0x98, 0x2c,
	0x26, 0x59, 0x49, 0x56, 0x8b, 0x9f, 0xb8, 0xdf, 0xa0, 0xe7, 0xb2, 0x13, 0x63, 0xdb, 0x63, 0x6f,
	0xff, 0xdf, 0xf0, 0x67, 0xe7, 0x37, 0x0b, 0xff, 0x2a, 0xdc, 0xe8, 0x2a, 0xc1, 0x2a, 0xda, 0x57,
	0xda, 0x68, 0xde, 0x6b, 0x79, 0x78, 0x9b, 0x6a, 0x9d, 0xe6, 0x38, 0xa6, 0xf9, 0xfa, 0xb0, 0x1d,
	0x1b, 0x55, 0x60, 0x6d, 0x56, 0xc5, 0xbe, 0xa9, 0x8e, 0xde, 0x19, 0x74, 0x1f, 0x5f, 0x16, 0x8b,
	0xe9, 0xf3, 0x8c, 0x73, 0x70, 0x73, 0x55, 0xa2, 0x60, 0x01, 0x0b, 0x7d, 0x49, 0x99, 0x5f, 0x41,
	0x67, 0xa3, 0x0f, 0xa5, 0x11, 0x7f, 0x02, 0x16, 0x0e, 0x64, 0x03, 0xfc, 0x01, 0xfc, 0xcb, 0x43,
	0xc2, 0x09, 0x58, 0xd8, 0x9f, 0x0c, 0xa3, 0x66, 0x55, 0xd4, 0xae, 0x8a, 0x5e, 0xdb, 0x86, 0xfc,
	0x2a, 0xf3, 0x7b, 0x80, 0x1d, 0x9e, 0x6a, 0x53, 0xe9, 0x1d, 0xd6, 0xc2, 0x0d, 0x9c, 0xb0, 0x3f,
	0xb9, 0x8e, 0x2e, 0xfe, 0x67, 0x95, 0x68, 0x1e, 0x2f, 0xe5, 0xb7, 0xe2, 0x70, 0x0a, 0xce, 0x3c,
	0x5e, 0xf2, 0x1b, 0xf0, 0xf4, 0x76, 0x5b, 0xa3, 0x21, 0x47, 0x26, 0xcf, 0xc4, 0xff, 0x83, 0x63,
	0x56, 0x29, 0x39, 0xfa, 0xd2, 0x46, 0x3b, 0xd9, 0xe1, 0x89, 0xdc, 0x7c, 0x69, 0xe3, 0xe8, 0x83,
	0x81, 0xf7, 0x14, 0x4f, 0x67, 0xb1, 0xe4, 0x02, 0xba, 0x47, 0xac, 0x6a, 0xa5, 0x4b, 0x7a, 0x67,
	0x20, 0x5b, 0xb4, 0xe7, 0xbe, 0xa9, 0xc4, 0x64, 0xed, 0xb9, 0x04, 0x76, 0x6d, 0x86, 0x2a, 0xcd,
	0x0c, 0xbd, 0x37, 0x90, 0x67, 0xfa, 0xf9, 0x0d, 0xee, 0x6f, 0xbe, 0x81, 0x83, 0xab, 0x92, 0x1c,
	0x45, 0x87, 0xce, 0xa0, 0x6c, 0xad, 0x36, 0xba, 0x28, 0x56, 0x65, 0x22, 0x3c, 0xd2, 0x6e, 0xd1,
	0x5a, 0x19, 0x65, 0x72, 0x14, 0x5d, 0x9a, 0x37, 0x60, 0x4f, 0xc4, 0xf2, 0x28, 0x7a, 0x01, 0x0b,
	0xff, 0x4a, 0x1b, 0xa9, 0x97, 0x61, 0x81, 0xc2, 0xa7, 0x59, 0x03, 0x6b, 0x8f, 0x54, 0xee, 0x3e,
	0x03, 0x00, 0x00, 0xff, 0xff, 0x4d, 0x86, 0x5b, 0xb8, 0x26, 0x02, 0x00, 0x00,
}
