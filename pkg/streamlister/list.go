package streamlister

import (
	"fmt"
	"io"
	"sync"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func unmarshalListStream(dAtA *streamBuffer, param ParamInterface) error {
	bufCh := make(chan []byte, 4000)
	errCh := make(chan error, 1)
	const parallelism = 2
	wg := &sync.WaitGroup{}
	wg.Add(parallelism)
	for i := 0; i < parallelism; i++ {
		go func() {
			defer wg.Done()
			for buf := range bufCh {
				obj := param.ObjectFactory()
				if err := obj.Unmarshal(buf); err != nil {
					select {
					case errCh <- err:
					default:
					}
					return
				} else {
					param.OnObject(obj)
				}
			}
		}()
	}

	l := dAtA.Len()
	iNdEx := 0
Loop:
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return corev1.ErrIntOverflowGenerated
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b, err := dAtA.Get(iNdEx)
			if err != nil {
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
			return fmt.Errorf("proto: PodList: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: PodList: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ListMeta", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return corev1.ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b, err := dAtA.Get(iNdEx)
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
				return corev1.ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return corev1.ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			buf, err := dAtA.Slice(iNdEx, postIndex)
			if err != nil {
				return err
			}
			var listMeta metav1.ListMeta
			if err := listMeta.Unmarshal(buf); err != nil {
				return err
			}
			param.OnListMeta(&listMeta)
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Items", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return corev1.ErrIntOverflowGenerated
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b, err := dAtA.Get(iNdEx)
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
				return corev1.ErrInvalidLengthGenerated
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return corev1.ErrInvalidLengthGenerated
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			buf, err := dAtA.Slice(iNdEx, postIndex)
			if err != nil {
				return err
			}

			select {
			case err = <-errCh:
				close(bufCh)
				wg.Wait()
				return err
			case bufCh <- buf:
			}

			iNdEx = postIndex
		default:
			if err := dAtA.Discard(); err != nil {
				return err
			}
			break Loop
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}

	close(bufCh)
	wg.Wait()
	return nil
}
