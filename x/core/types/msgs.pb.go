// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: core/msgs.proto

package types

import (
	fmt "fmt"
	types1 "github.com/cosmos/cosmos-sdk/codec/types"
	types "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type MsgInitiate struct {
	Sender         string `protobuf:"bytes,1,opt,name=sender,proto3" json:"sender,omitempty"`
	Nonce          uint64 `protobuf:"varint,2,opt,name=nonce,proto3" json:"nonce,omitempty"`
	CommitProtocol uint32 `protobuf:"varint,3,opt,name=commit_protocol,json=commitProtocol,proto3" json:"commit_protocol,omitempty"`
	// Timeout height relative to the current block height.
	// The timeout is disabled when set to 0.
	TimeoutHeight types.Height `protobuf:"bytes,4,opt,name=timeout_height,json=timeoutHeight,proto3" json:"timeout_height" yaml:"timeout_height"`
	// Timeout timestamp (in nanoseconds) relative to the current block timestamp.
	// The timeout is disabled when set to 0.
	TimeoutTimestamp     uint64                `protobuf:"varint,5,opt,name=timeout_timestamp,json=timeoutTimestamp,proto3" json:"timeout_timestamp,omitempty" yaml:"timeout_timestamp"`
	ContractTransactions []ContractTransaction `protobuf:"bytes,6,rep,name=contract_transactions,json=contractTransactions,proto3" json:"contract_transactions"`
}

func (m *MsgInitiate) Reset()         { *m = MsgInitiate{} }
func (m *MsgInitiate) String() string { return proto.CompactTextString(m) }
func (*MsgInitiate) ProtoMessage()    {}
func (*MsgInitiate) Descriptor() ([]byte, []int) {
	return fileDescriptor_cb9cfdd03571ef08, []int{0}
}
func (m *MsgInitiate) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MsgInitiate) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MsgInitiate.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MsgInitiate) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MsgInitiate.Merge(m, src)
}
func (m *MsgInitiate) XXX_Size() int {
	return m.Size()
}
func (m *MsgInitiate) XXX_DiscardUnknown() {
	xxx_messageInfo_MsgInitiate.DiscardUnknown(m)
}

var xxx_messageInfo_MsgInitiate proto.InternalMessageInfo

type ContractTransaction struct {
	ChainId         types1.Any      `protobuf:"bytes,1,opt,name=chain_id,json=chainId,proto3" json:"chain_id"`
	Signers         []string        `protobuf:"bytes,2,rep,name=signers,proto3" json:"signers,omitempty"`
	CallInfo        []byte          `protobuf:"bytes,3,opt,name=call_info,json=callInfo,proto3" json:"call_info,omitempty"`
	StateConstraint StateConstraint `protobuf:"bytes,4,opt,name=state_constraint,json=stateConstraint,proto3" json:"state_constraint"`
	ReturnValue     *ReturnValue    `protobuf:"bytes,5,opt,name=return_value,json=returnValue,proto3" json:"return_value,omitempty"`
	Links           []types1.Any    `protobuf:"bytes,6,rep,name=links,proto3" json:"links"`
}

func (m *ContractTransaction) Reset()         { *m = ContractTransaction{} }
func (m *ContractTransaction) String() string { return proto.CompactTextString(m) }
func (*ContractTransaction) ProtoMessage()    {}
func (*ContractTransaction) Descriptor() ([]byte, []int) {
	return fileDescriptor_cb9cfdd03571ef08, []int{1}
}
func (m *ContractTransaction) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ContractTransaction) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ContractTransaction.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ContractTransaction) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ContractTransaction.Merge(m, src)
}
func (m *ContractTransaction) XXX_Size() int {
	return m.Size()
}
func (m *ContractTransaction) XXX_DiscardUnknown() {
	xxx_messageInfo_ContractTransaction.DiscardUnknown(m)
}

var xxx_messageInfo_ContractTransaction proto.InternalMessageInfo

type StateConstraint struct {
	Type uint32       `protobuf:"varint,1,opt,name=type,proto3" json:"type,omitempty"`
	Ops  []types1.Any `protobuf:"bytes,2,rep,name=ops,proto3" json:"ops"`
}

func (m *StateConstraint) Reset()         { *m = StateConstraint{} }
func (m *StateConstraint) String() string { return proto.CompactTextString(m) }
func (*StateConstraint) ProtoMessage()    {}
func (*StateConstraint) Descriptor() ([]byte, []int) {
	return fileDescriptor_cb9cfdd03571ef08, []int{2}
}
func (m *StateConstraint) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *StateConstraint) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_StateConstraint.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *StateConstraint) XXX_Merge(src proto.Message) {
	xxx_messageInfo_StateConstraint.Merge(m, src)
}
func (m *StateConstraint) XXX_Size() int {
	return m.Size()
}
func (m *StateConstraint) XXX_DiscardUnknown() {
	xxx_messageInfo_StateConstraint.DiscardUnknown(m)
}

var xxx_messageInfo_StateConstraint proto.InternalMessageInfo

type ReturnValue struct {
	Value []byte `protobuf:"bytes,1,opt,name=value,proto3" json:"value,omitempty"`
}

func (m *ReturnValue) Reset()         { *m = ReturnValue{} }
func (m *ReturnValue) String() string { return proto.CompactTextString(m) }
func (*ReturnValue) ProtoMessage()    {}
func (*ReturnValue) Descriptor() ([]byte, []int) {
	return fileDescriptor_cb9cfdd03571ef08, []int{3}
}
func (m *ReturnValue) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ReturnValue) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ReturnValue.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ReturnValue) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReturnValue.Merge(m, src)
}
func (m *ReturnValue) XXX_Size() int {
	return m.Size()
}
func (m *ReturnValue) XXX_DiscardUnknown() {
	xxx_messageInfo_ReturnValue.DiscardUnknown(m)
}

var xxx_messageInfo_ReturnValue proto.InternalMessageInfo

func init() {
	proto.RegisterType((*MsgInitiate)(nil), "cross.core.MsgInitiate")
	proto.RegisterType((*ContractTransaction)(nil), "cross.core.ContractTransaction")
	proto.RegisterType((*StateConstraint)(nil), "cross.core.StateConstraint")
	proto.RegisterType((*ReturnValue)(nil), "cross.core.ReturnValue")
}

func init() { proto.RegisterFile("core/msgs.proto", fileDescriptor_cb9cfdd03571ef08) }

var fileDescriptor_cb9cfdd03571ef08 = []byte{
	// 598 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x93, 0x4d, 0x6f, 0xd3, 0x4c,
	0x10, 0xc7, 0xed, 0x24, 0x7d, 0x5b, 0xf7, 0xed, 0xd9, 0x27, 0x05, 0xd3, 0x82, 0x6d, 0xe5, 0x82,
	0x85, 0xc0, 0xa6, 0x41, 0x5c, 0x7a, 0x23, 0xe5, 0x40, 0x24, 0x90, 0x90, 0xa9, 0x38, 0x54, 0x42,
	0x66, 0xbd, 0xd9, 0x3a, 0x2b, 0xec, 0xdd, 0xc8, 0xbb, 0xa9, 0xc8, 0x37, 0xe0, 0xc8, 0x47, 0xe8,
	0xc7, 0x29, 0xb7, 0x1e, 0x39, 0x55, 0xa8, 0xb9, 0x70, 0xee, 0x85, 0x2b, 0xda, 0xb5, 0xdd, 0x24,
	0x80, 0x7a, 0xf2, 0xce, 0x7f, 0x7e, 0x3b, 0xfb, 0x9f, 0x19, 0x19, 0x6c, 0x61, 0x5e, 0x90, 0x30,
	0x17, 0xa9, 0x08, 0x46, 0x05, 0x97, 0x1c, 0x02, 0x5c, 0x70, 0x21, 0x02, 0x25, 0xef, 0xb6, 0x53,
	0x9e, 0x72, 0x2d, 0x87, 0xea, 0x54, 0x12, 0xbb, 0x2e, 0x4d, 0x70, 0xa8, 0xaf, 0xe1, 0x8c, 0x12,
	0x26, 0xc3, 0xd3, 0xfd, 0xea, 0x54, 0x01, 0xf7, 0x52, 0xce, 0xd3, 0x8c, 0x84, 0x3a, 0x4a, 0xc6,
	0x27, 0x21, 0x62, 0x93, 0x32, 0xd5, 0xf9, 0xd5, 0x00, 0xd6, 0x1b, 0x91, 0xf6, 0x19, 0x95, 0x14,
	0x49, 0x02, 0xef, 0x80, 0x65, 0x41, 0xd8, 0x80, 0x14, 0xb6, 0xe9, 0x99, 0xfe, 0x5a, 0x54, 0x45,
	0xb0, 0x0d, 0x96, 0x18, 0x67, 0x98, 0xd8, 0x0d, 0xcf, 0xf4, 0x5b, 0x51, 0x19, 0xc0, 0x87, 0xca,
	0x6e, 0x9e, 0x53, 0x19, 0xeb, 0x6a, 0x98, 0x67, 0x76, 0xd3, 0x33, 0xfd, 0x8d, 0x68, 0xb3, 0x94,
	0xdf, 0x56, 0x2a, 0xfc, 0x08, 0x36, 0x25, 0xcd, 0x09, 0x1f, 0xcb, 0x78, 0x48, 0x68, 0x3a, 0x94,
	0x76, 0xcb, 0x33, 0x7d, 0xab, 0xbb, 0x1b, 0xd0, 0x04, 0xeb, 0xde, 0x82, 0xca, 0xf1, 0xe9, 0x7e,
	0xf0, 0x4a, 0x13, 0xbd, 0x07, 0xe7, 0x97, 0xae, 0x71, 0x7d, 0xe9, 0xee, 0x4c, 0x50, 0x9e, 0x1d,
	0x74, 0x16, 0xef, 0x77, 0xa2, 0x8d, 0x4a, 0x28, 0x69, 0xd8, 0x07, 0xff, 0xd5, 0x84, 0xfa, 0x0a,
	0x89, 0xf2, 0x91, 0xbd, 0xa4, 0xcc, 0xf6, 0xee, 0x5f, 0x5f, 0xba, 0xf6, 0x62, 0x91, 0x1b, 0xa4,
	0x13, 0x6d, 0x57, 0xda, 0x51, 0x2d, 0xc1, 0x63, 0xb0, 0x83, 0x39, 0x93, 0x05, 0xc2, 0x32, 0x96,
	0x05, 0x62, 0x02, 0x61, 0x49, 0x39, 0x13, 0xf6, 0xb2, 0xd7, 0xf4, 0xad, 0xae, 0x1b, 0xcc, 0x36,
	0x12, 0x1c, 0x56, 0xe0, 0xd1, 0x8c, 0xeb, 0xb5, 0x94, 0xf1, 0xa8, 0x8d, 0xff, 0x4e, 0x89, 0x83,
	0xd5, 0x2f, 0x67, 0xae, 0xf1, 0xf3, 0xcc, 0x35, 0x3a, 0xdf, 0x1a, 0xe0, 0xff, 0x7f, 0xdc, 0x86,
	0xcf, 0xc1, 0x2a, 0x1e, 0x22, 0xca, 0x62, 0x3a, 0xd0, 0x3b, 0xb0, 0xba, 0xed, 0xa0, 0xdc, 0x5f,
	0x50, 0xef, 0x2f, 0x78, 0xc1, 0x26, 0xd5, 0x2b, 0x2b, 0x9a, 0xed, 0x0f, 0xa0, 0x0d, 0x56, 0x04,
	0x4d, 0x19, 0x29, 0x84, 0xdd, 0xf0, 0x9a, 0xfe, 0x5a, 0x54, 0x87, 0x70, 0x0f, 0xac, 0x61, 0x94,
	0x65, 0x31, 0x65, 0x27, 0x5c, 0xaf, 0x67, 0x3d, 0x5a, 0x55, 0x42, 0x9f, 0x9d, 0x70, 0xf8, 0x1a,
	0x6c, 0x0b, 0x89, 0x24, 0x89, 0x31, 0x67, 0x42, 0x16, 0x88, 0xb2, 0x7a, 0x35, 0x7b, 0xf3, 0x6d,
	0xbe, 0x53, 0xcc, 0xe1, 0x0d, 0x52, 0x3d, 0xbe, 0x25, 0x16, 0x65, 0x78, 0x00, 0xd6, 0x0b, 0x22,
	0xc7, 0x05, 0x8b, 0x4f, 0x51, 0x36, 0x26, 0x7a, 0xfe, 0x56, 0xf7, 0xee, 0x7c, 0xa5, 0x48, 0xe7,
	0xdf, 0xab, 0x74, 0x64, 0x15, 0xb3, 0x00, 0x3e, 0x05, 0x4b, 0x19, 0x65, 0x9f, 0xea, 0x29, 0xdf,
	0xd6, 0x74, 0x09, 0xce, 0xcd, 0xf2, 0x03, 0xd8, 0xfa, 0xc3, 0x21, 0x84, 0xa0, 0x25, 0x27, 0x23,
	0xa2, 0x47, 0xb8, 0x11, 0xe9, 0x33, 0x7c, 0x0c, 0x9a, 0x7c, 0x54, 0xce, 0xe7, 0xf6, 0x07, 0x14,
	0x36, 0x57, 0xfe, 0x09, 0xb0, 0xe6, 0x6c, 0xab, 0x7f, 0xa1, 0x6c, 0xcf, 0xd4, 0xc3, 0x2c, 0x83,
	0x19, 0xde, 0x7b, 0x79, 0x7e, 0xe5, 0x98, 0x17, 0x57, 0x8e, 0xf9, 0xe3, 0xca, 0x31, 0xbf, 0x4e,
	0x1d, 0xe3, 0x62, 0xea, 0x18, 0xdf, 0xa7, 0x8e, 0x71, 0xfc, 0x28, 0xa5, 0x72, 0x38, 0x4e, 0x02,
	0xcc, 0xf3, 0x70, 0x80, 0x24, 0xd2, 0xcb, 0xcb, 0x50, 0x12, 0xea, 0x01, 0x85, 0x9f, 0xcb, 0xbf,
	0x58, 0x79, 0x15, 0xc9, 0xb2, 0xf6, 0xf5, 0xec, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0xbf, 0x27,
	0x03, 0x3f, 0x11, 0x04, 0x00, 0x00,
}

func (m *MsgInitiate) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MsgInitiate) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MsgInitiate) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ContractTransactions) > 0 {
		for iNdEx := len(m.ContractTransactions) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.ContractTransactions[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintMsgs(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x32
		}
	}
	if m.TimeoutTimestamp != 0 {
		i = encodeVarintMsgs(dAtA, i, uint64(m.TimeoutTimestamp))
		i--
		dAtA[i] = 0x28
	}
	{
		size, err := m.TimeoutHeight.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintMsgs(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x22
	if m.CommitProtocol != 0 {
		i = encodeVarintMsgs(dAtA, i, uint64(m.CommitProtocol))
		i--
		dAtA[i] = 0x18
	}
	if m.Nonce != 0 {
		i = encodeVarintMsgs(dAtA, i, uint64(m.Nonce))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Sender) > 0 {
		i -= len(m.Sender)
		copy(dAtA[i:], m.Sender)
		i = encodeVarintMsgs(dAtA, i, uint64(len(m.Sender)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *ContractTransaction) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ContractTransaction) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ContractTransaction) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Links) > 0 {
		for iNdEx := len(m.Links) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Links[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintMsgs(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x32
		}
	}
	if m.ReturnValue != nil {
		{
			size, err := m.ReturnValue.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintMsgs(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x2a
	}
	{
		size, err := m.StateConstraint.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintMsgs(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x22
	if len(m.CallInfo) > 0 {
		i -= len(m.CallInfo)
		copy(dAtA[i:], m.CallInfo)
		i = encodeVarintMsgs(dAtA, i, uint64(len(m.CallInfo)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Signers) > 0 {
		for iNdEx := len(m.Signers) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.Signers[iNdEx])
			copy(dAtA[i:], m.Signers[iNdEx])
			i = encodeVarintMsgs(dAtA, i, uint64(len(m.Signers[iNdEx])))
			i--
			dAtA[i] = 0x12
		}
	}
	{
		size, err := m.ChainId.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintMsgs(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func (m *StateConstraint) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *StateConstraint) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *StateConstraint) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Ops) > 0 {
		for iNdEx := len(m.Ops) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.Ops[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintMsgs(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0x12
		}
	}
	if m.Type != 0 {
		i = encodeVarintMsgs(dAtA, i, uint64(m.Type))
		i--
		dAtA[i] = 0x8
	}
	return len(dAtA) - i, nil
}

func (m *ReturnValue) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ReturnValue) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ReturnValue) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Value) > 0 {
		i -= len(m.Value)
		copy(dAtA[i:], m.Value)
		i = encodeVarintMsgs(dAtA, i, uint64(len(m.Value)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintMsgs(dAtA []byte, offset int, v uint64) int {
	offset -= sovMsgs(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *MsgInitiate) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Sender)
	if l > 0 {
		n += 1 + l + sovMsgs(uint64(l))
	}
	if m.Nonce != 0 {
		n += 1 + sovMsgs(uint64(m.Nonce))
	}
	if m.CommitProtocol != 0 {
		n += 1 + sovMsgs(uint64(m.CommitProtocol))
	}
	l = m.TimeoutHeight.Size()
	n += 1 + l + sovMsgs(uint64(l))
	if m.TimeoutTimestamp != 0 {
		n += 1 + sovMsgs(uint64(m.TimeoutTimestamp))
	}
	if len(m.ContractTransactions) > 0 {
		for _, e := range m.ContractTransactions {
			l = e.Size()
			n += 1 + l + sovMsgs(uint64(l))
		}
	}
	return n
}

func (m *ContractTransaction) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.ChainId.Size()
	n += 1 + l + sovMsgs(uint64(l))
	if len(m.Signers) > 0 {
		for _, s := range m.Signers {
			l = len(s)
			n += 1 + l + sovMsgs(uint64(l))
		}
	}
	l = len(m.CallInfo)
	if l > 0 {
		n += 1 + l + sovMsgs(uint64(l))
	}
	l = m.StateConstraint.Size()
	n += 1 + l + sovMsgs(uint64(l))
	if m.ReturnValue != nil {
		l = m.ReturnValue.Size()
		n += 1 + l + sovMsgs(uint64(l))
	}
	if len(m.Links) > 0 {
		for _, e := range m.Links {
			l = e.Size()
			n += 1 + l + sovMsgs(uint64(l))
		}
	}
	return n
}

func (m *StateConstraint) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Type != 0 {
		n += 1 + sovMsgs(uint64(m.Type))
	}
	if len(m.Ops) > 0 {
		for _, e := range m.Ops {
			l = e.Size()
			n += 1 + l + sovMsgs(uint64(l))
		}
	}
	return n
}

func (m *ReturnValue) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Value)
	if l > 0 {
		n += 1 + l + sovMsgs(uint64(l))
	}
	return n
}

func sovMsgs(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozMsgs(x uint64) (n int) {
	return sovMsgs(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *MsgInitiate) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMsgs
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: MsgInitiate: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MsgInitiate: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Sender", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgs
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMsgs
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMsgs
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Sender = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Nonce", wireType)
			}
			m.Nonce = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgs
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Nonce |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field CommitProtocol", wireType)
			}
			m.CommitProtocol = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgs
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.CommitProtocol |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TimeoutHeight", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgs
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMsgs
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMsgs
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.TimeoutHeight.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field TimeoutTimestamp", wireType)
			}
			m.TimeoutTimestamp = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgs
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.TimeoutTimestamp |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContractTransactions", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgs
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMsgs
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMsgs
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ContractTransactions = append(m.ContractTransactions, ContractTransaction{})
			if err := m.ContractTransactions[len(m.ContractTransactions)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMsgs(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMsgs
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthMsgs
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *ContractTransaction) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMsgs
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ContractTransaction: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ContractTransaction: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ChainId", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgs
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMsgs
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMsgs
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.ChainId.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Signers", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgs
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthMsgs
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMsgs
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Signers = append(m.Signers, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CallInfo", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgs
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthMsgs
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthMsgs
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.CallInfo = append(m.CallInfo[:0], dAtA[iNdEx:postIndex]...)
			if m.CallInfo == nil {
				m.CallInfo = []byte{}
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field StateConstraint", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgs
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMsgs
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMsgs
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.StateConstraint.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ReturnValue", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgs
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMsgs
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMsgs
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.ReturnValue == nil {
				m.ReturnValue = &ReturnValue{}
			}
			if err := m.ReturnValue.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Links", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgs
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMsgs
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMsgs
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Links = append(m.Links, types1.Any{})
			if err := m.Links[len(m.Links)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMsgs(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMsgs
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthMsgs
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *StateConstraint) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMsgs
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: StateConstraint: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: StateConstraint: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			m.Type = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgs
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.Type |= uint32(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Ops", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgs
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthMsgs
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthMsgs
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Ops = append(m.Ops, types1.Any{})
			if err := m.Ops[len(m.Ops)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMsgs(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMsgs
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthMsgs
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *ReturnValue) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMsgs
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: ReturnValue: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ReturnValue: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Value", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMsgs
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthMsgs
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthMsgs
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Value = append(m.Value[:0], dAtA[iNdEx:postIndex]...)
			if m.Value == nil {
				m.Value = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMsgs(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMsgs
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthMsgs
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipMsgs(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMsgs
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMsgs
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowMsgs
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthMsgs
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupMsgs
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthMsgs
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthMsgs        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMsgs          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupMsgs = fmt.Errorf("proto: unexpected end of group")
)