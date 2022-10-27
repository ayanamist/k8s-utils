package streamlister

import (
	"errors"
	"fmt"
	"io"

	"k8s.io/apimachinery/pkg/runtime"
)

type unknownStreamUnmarshaler struct {
	OnTypeMeta        func(runtime.TypeMeta)
	OnRaw             func(*streamBuffer) error
	OnContentEncoding func(string)
	OnContentType     func(string)
}

func (u unknownStreamUnmarshaler) Unmarshal(buffer *streamBuffer) error {
	onTypeMeta := u.OnTypeMeta
	if onTypeMeta == nil {
		onTypeMeta = func(runtime.TypeMeta) {}
	}
	onRaw := u.OnRaw
	if onRaw == nil {
		onRaw = func(b *streamBuffer) error {
			return b.Discard()
		}
	}
	onContentEncoding := u.OnContentEncoding
	if onContentEncoding == nil {
		onContentEncoding = func(string) {}
	}
	onContentType := u.OnContentType
	if onContentType == nil {
		onContentType = func(string) {}
	}

	iNdEx := 0
Loop:
	for {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return runtime.ErrIntOverflowGenerated
			}
			b, err := buffer.Get(iNdEx)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break Loop
				}
				return err
			}
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: Unknown: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Unknown: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TypeMeta", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return runtime.ErrIntOverflowGenerated
				}
				b, err := buffer.Get(iNdEx)
				if err != nil {
					return err
				}
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return runtime.ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return runtime.ErrInvalidLengthGenerated
			}
			buf, err := buffer.Slice(iNdEx, postIndex)
			if err != nil {
				return err
			}
			var typeMeta runtime.TypeMeta
			if err := typeMeta.Unmarshal(buf); err != nil {
				return err
			}
			onTypeMeta(typeMeta)
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Raw", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return runtime.ErrIntOverflowGenerated
				}
				b, err := buffer.Get(iNdEx)
				if err != nil {
					return err
				}
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return runtime.ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return runtime.ErrInvalidLengthGenerated
			}
			stream, err := buffer.SubStream(iNdEx, postIndex)
			if err != nil {
				return err
			}
			if err := onRaw(stream); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContentEncoding", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return runtime.ErrIntOverflowGenerated
				}
				b, err := buffer.Get(iNdEx)
				if err != nil {
					return err
				}
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return runtime.ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return runtime.ErrInvalidLengthGenerated
			}
			buf, err := buffer.Slice(iNdEx, postIndex)
			if err != nil {
				return err
			}
			onContentEncoding(string(buf))
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ContentType", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return runtime.ErrIntOverflowGenerated
				}
				b, err := buffer.Get(iNdEx)
				if err != nil {
					return err
				}
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return runtime.ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return runtime.ErrInvalidLengthGenerated
			}
			buf, err := buffer.Slice(iNdEx, postIndex)
			if err != nil {
				return err
			}
			onContentType(string(buf))
			iNdEx = postIndex
		default:
			break Loop
		}
	}
	return buffer.Discard()
}
