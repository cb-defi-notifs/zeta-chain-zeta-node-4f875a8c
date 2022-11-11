// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: observer/observer.proto

package types

import (
	fmt "fmt"
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

type ObserverChain int32

const (
	ObserverChain_Empty      ObserverChain = 0
	ObserverChain_Eth        ObserverChain = 1
	ObserverChain_ZetaChain  ObserverChain = 2
	ObserverChain_Btc        ObserverChain = 3
	ObserverChain_Polygon    ObserverChain = 4
	ObserverChain_BscMainnet ObserverChain = 5
	ObserverChain_Goerli     ObserverChain = 6
	ObserverChain_Mumbai     ObserverChain = 7
	ObserverChain_Ropsten    ObserverChain = 8
	ObserverChain_Ganache    ObserverChain = 9
	ObserverChain_Baobap     ObserverChain = 10
	ObserverChain_BscTestnet ObserverChain = 11
)

var ObserverChain_name = map[int32]string{
	0:  "Empty",
	1:  "Eth",
	2:  "ZetaChain",
	3:  "Btc",
	4:  "Polygon",
	5:  "BscMainnet",
	6:  "Goerli",
	7:  "Mumbai",
	8:  "Ropsten",
	9:  "Ganache",
	10: "Baobap",
	11: "BscTestnet",
}

var ObserverChain_value = map[string]int32{
	"Empty":      0,
	"Eth":        1,
	"ZetaChain":  2,
	"Btc":        3,
	"Polygon":    4,
	"BscMainnet": 5,
	"Goerli":     6,
	"Mumbai":     7,
	"Ropsten":    8,
	"Ganache":    9,
	"Baobap":     10,
	"BscTestnet": 11,
}

func (x ObserverChain) String() string {
	return proto.EnumName(ObserverChain_name, int32(x))
}

func (ObserverChain) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_3004233a4a5969ce, []int{0}
}

type ObservationType int32

const (
	ObservationType_EmptyObserverType ObservationType = 0
	ObservationType_InBoundTx         ObservationType = 1
	ObservationType_OutBoundTx        ObservationType = 2
	ObservationType_GasPrice          ObservationType = 3
)

var ObservationType_name = map[int32]string{
	0: "EmptyObserverType",
	1: "InBoundTx",
	2: "OutBoundTx",
	3: "GasPrice",
}

var ObservationType_value = map[string]int32{
	"EmptyObserverType": 0,
	"InBoundTx":         1,
	"OutBoundTx":        2,
	"GasPrice":          3,
}

func (x ObservationType) String() string {
	return proto.EnumName(ObservationType_name, int32(x))
}

func (ObservationType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_3004233a4a5969ce, []int{1}
}

type ObserverMapper struct {
	Index           string          `protobuf:"bytes,1,opt,name=index,proto3" json:"index,omitempty"`
	ObserverChain   ObserverChain   `protobuf:"varint,2,opt,name=ObserverChain,proto3,enum=zetachain.zetacore.observer.ObserverChain" json:"ObserverChain,omitempty"`
	ObservationType ObservationType `protobuf:"varint,3,opt,name=ObservationType,proto3,enum=zetachain.zetacore.observer.ObservationType" json:"ObservationType,omitempty"`
	ObserverList    []string        `protobuf:"bytes,4,rep,name=observerList,proto3" json:"observerList,omitempty"`
}

func (m *ObserverMapper) Reset()         { *m = ObserverMapper{} }
func (m *ObserverMapper) String() string { return proto.CompactTextString(m) }
func (*ObserverMapper) ProtoMessage()    {}
func (*ObserverMapper) Descriptor() ([]byte, []int) {
	return fileDescriptor_3004233a4a5969ce, []int{0}
}
func (m *ObserverMapper) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ObserverMapper) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ObserverMapper.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ObserverMapper) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ObserverMapper.Merge(m, src)
}
func (m *ObserverMapper) XXX_Size() int {
	return m.Size()
}
func (m *ObserverMapper) XXX_DiscardUnknown() {
	xxx_messageInfo_ObserverMapper.DiscardUnknown(m)
}

var xxx_messageInfo_ObserverMapper proto.InternalMessageInfo

func (m *ObserverMapper) GetIndex() string {
	if m != nil {
		return m.Index
	}
	return ""
}

func (m *ObserverMapper) GetObserverChain() ObserverChain {
	if m != nil {
		return m.ObserverChain
	}
	return ObserverChain_Empty
}

func (m *ObserverMapper) GetObservationType() ObservationType {
	if m != nil {
		return m.ObservationType
	}
	return ObservationType_EmptyObserverType
}

func (m *ObserverMapper) GetObserverList() []string {
	if m != nil {
		return m.ObserverList
	}
	return nil
}

func init() {
	proto.RegisterEnum("zetachain.zetacore.observer.ObserverChain", ObserverChain_name, ObserverChain_value)
	proto.RegisterEnum("zetachain.zetacore.observer.ObservationType", ObservationType_name, ObservationType_value)
	proto.RegisterType((*ObserverMapper)(nil), "zetachain.zetacore.observer.ObserverMapper")
}

func init() { proto.RegisterFile("observer/observer.proto", fileDescriptor_3004233a4a5969ce) }

var fileDescriptor_3004233a4a5969ce = []byte{
	// 425 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x92, 0x4f, 0x6a, 0xdb, 0x40,
	0x14, 0xc6, 0x35, 0x96, 0xff, 0x44, 0x2f, 0x89, 0x3b, 0x1d, 0x52, 0x2a, 0x52, 0x10, 0x26, 0x2b,
	0x63, 0x5a, 0x09, 0xda, 0x1b, 0xa8, 0x04, 0x13, 0xa8, 0x89, 0x11, 0xa6, 0x8b, 0x74, 0x35, 0x52,
	0x06, 0x69, 0xc0, 0x9e, 0x19, 0xa4, 0x51, 0xb1, 0x7a, 0x8a, 0x1e, 0xa2, 0x8b, 0x2e, 0x7a, 0x90,
	0x2e, 0xbd, 0xec, 0xb2, 0xd8, 0x57, 0xe8, 0x01, 0xca, 0x48, 0x96, 0xc1, 0x2e, 0x94, 0xec, 0xde,
	0xfb, 0xf8, 0xbe, 0xdf, 0xfc, 0x79, 0x0f, 0x5e, 0xca, 0xb8, 0x60, 0xf9, 0x67, 0x96, 0x07, 0x6d,
	0xe1, 0xab, 0x5c, 0x6a, 0x49, 0x5e, 0x7d, 0x61, 0x9a, 0x26, 0x19, 0xe5, 0xc2, 0xaf, 0x2b, 0x99,
	0x33, 0xbf, 0xb5, 0x5c, 0x5f, 0xa5, 0x32, 0x95, 0xb5, 0x2f, 0x30, 0x55, 0x13, 0xb9, 0xf9, 0x83,
	0x60, 0x78, 0xbf, 0xb7, 0xcc, 0xa8, 0x52, 0x2c, 0x27, 0x57, 0xd0, 0xe3, 0xe2, 0x91, 0xad, 0x5d,
	0x34, 0x42, 0x63, 0x27, 0x6a, 0x1a, 0x32, 0x87, 0xcb, 0xd6, 0xf7, 0xde, 0x9c, 0xe0, 0x76, 0x46,
	0x68, 0x3c, 0x7c, 0x3b, 0xf1, 0xff, 0x73, 0xa6, 0x7f, 0x94, 0x88, 0x8e, 0x01, 0xe4, 0x23, 0x3c,
	0x6b, 0x04, 0xaa, 0xb9, 0x14, 0x8b, 0x4a, 0x31, 0xd7, 0xae, 0x99, 0xaf, 0x9f, 0xc0, 0x3c, 0x64,
	0xa2, 0x53, 0x08, 0xb9, 0x81, 0x8b, 0xd6, 0xfc, 0x81, 0x17, 0xda, 0xed, 0x8e, 0xec, 0xb1, 0x13,
	0x1d, 0x69, 0x93, 0x1f, 0xe8, 0xe4, 0x39, 0xc4, 0x81, 0xde, 0xed, 0x4a, 0xe9, 0x0a, 0x5b, 0x64,
	0x00, 0xf6, 0xad, 0xce, 0x30, 0x22, 0x97, 0xe0, 0x3c, 0x30, 0x4d, 0x6b, 0x03, 0xee, 0x18, 0x3d,
	0xd4, 0x09, 0xb6, 0xc9, 0x39, 0x0c, 0xe6, 0x72, 0x59, 0xa5, 0x52, 0xe0, 0x2e, 0x19, 0x02, 0x84,
	0x45, 0x32, 0xa3, 0x5c, 0x08, 0xa6, 0x71, 0x8f, 0x00, 0xf4, 0xa7, 0x92, 0xe5, 0x4b, 0x8e, 0xfb,
	0xa6, 0x9e, 0x95, 0xab, 0x98, 0x72, 0x3c, 0x30, 0xa1, 0x48, 0xaa, 0x42, 0x33, 0x81, 0xcf, 0x4c,
	0x33, 0xa5, 0x82, 0x26, 0x19, 0xc3, 0x8e, 0x71, 0x85, 0x54, 0xc6, 0x54, 0x61, 0xd8, 0xd3, 0x16,
	0xac, 0xd0, 0x86, 0x76, 0x7e, 0xdd, 0xfd, 0xfe, 0xcd, 0x43, 0x93, 0x4f, 0xff, 0x7c, 0x15, 0x79,
	0x01, 0xcf, 0xeb, 0xfb, 0xb6, 0xaf, 0x30, 0x22, 0xb6, 0xcc, 0x95, 0xef, 0x44, 0x28, 0x4b, 0xf1,
	0xb8, 0x58, 0x63, 0x64, 0x70, 0xf7, 0xa5, 0x6e, 0xfb, 0x0e, 0xb9, 0x80, 0xb3, 0x29, 0x2d, 0xe6,
	0x39, 0x4f, 0x18, 0xb6, 0x1b, 0x78, 0x78, 0xf7, 0x73, 0xeb, 0xa1, 0xcd, 0xd6, 0x43, 0xbf, 0xb7,
	0x1e, 0xfa, 0xba, 0xf3, 0xac, 0xcd, 0xce, 0xb3, 0x7e, 0xed, 0x3c, 0xeb, 0x21, 0x48, 0xb9, 0xce,
	0xca, 0xd8, 0x4f, 0xe4, 0x2a, 0x30, 0x83, 0x78, 0x53, 0xcf, 0x24, 0x68, 0x67, 0x12, 0xac, 0x0f,
	0x0b, 0x18, 0xe8, 0x4a, 0xb1, 0x22, 0xee, 0xd7, 0x4b, 0xf5, 0xee, 0x6f, 0x00, 0x00, 0x00, 0xff,
	0xff, 0x7f, 0x93, 0x67, 0xfd, 0xa2, 0x02, 0x00, 0x00,
}

func (m *ObserverMapper) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ObserverMapper) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ObserverMapper) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.ObserverList) > 0 {
		for iNdEx := len(m.ObserverList) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.ObserverList[iNdEx])
			copy(dAtA[i:], m.ObserverList[iNdEx])
			i = encodeVarintObserver(dAtA, i, uint64(len(m.ObserverList[iNdEx])))
			i--
			dAtA[i] = 0x22
		}
	}
	if m.ObservationType != 0 {
		i = encodeVarintObserver(dAtA, i, uint64(m.ObservationType))
		i--
		dAtA[i] = 0x18
	}
	if m.ObserverChain != 0 {
		i = encodeVarintObserver(dAtA, i, uint64(m.ObserverChain))
		i--
		dAtA[i] = 0x10
	}
	if len(m.Index) > 0 {
		i -= len(m.Index)
		copy(dAtA[i:], m.Index)
		i = encodeVarintObserver(dAtA, i, uint64(len(m.Index)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintObserver(dAtA []byte, offset int, v uint64) int {
	offset -= sovObserver(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *ObserverMapper) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Index)
	if l > 0 {
		n += 1 + l + sovObserver(uint64(l))
	}
	if m.ObserverChain != 0 {
		n += 1 + sovObserver(uint64(m.ObserverChain))
	}
	if m.ObservationType != 0 {
		n += 1 + sovObserver(uint64(m.ObservationType))
	}
	if len(m.ObserverList) > 0 {
		for _, s := range m.ObserverList {
			l = len(s)
			n += 1 + l + sovObserver(uint64(l))
		}
	}
	return n
}

func sovObserver(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozObserver(x uint64) (n int) {
	return sovObserver(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *ObserverMapper) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowObserver
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
			return fmt.Errorf("proto: ObserverMapper: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ObserverMapper: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Index", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowObserver
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
				return ErrInvalidLengthObserver
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthObserver
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Index = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ObserverChain", wireType)
			}
			m.ObserverChain = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowObserver
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ObserverChain |= ObserverChain(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 3:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field ObservationType", wireType)
			}
			m.ObservationType = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowObserver
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.ObservationType |= ObservationType(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ObserverList", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowObserver
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
				return ErrInvalidLengthObserver
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthObserver
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.ObserverList = append(m.ObserverList, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipObserver(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthObserver
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
func skipObserver(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowObserver
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
					return 0, ErrIntOverflowObserver
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
					return 0, ErrIntOverflowObserver
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
				return 0, ErrInvalidLengthObserver
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupObserver
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthObserver
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthObserver        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowObserver          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupObserver = fmt.Errorf("proto: unexpected end of group")
)