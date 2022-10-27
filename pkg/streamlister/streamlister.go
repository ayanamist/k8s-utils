package streamlister

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/gogo/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type ParamInterface interface {
	ObjectFactory() proto.Unmarshaler
	OnListMeta(*metav1.ListMeta)
	OnObject(proto.Unmarshaler)
}

type Param struct {
	ObjectFactoryFunc func() proto.Unmarshaler
	OnListMetaFunc    func(*metav1.ListMeta)
	OnObjectFunc      func(proto.Unmarshaler)
}

func (p Param) ObjectFactory() proto.Unmarshaler {
	return p.ObjectFactoryFunc()
}

func (p Param) OnListMeta(meta *metav1.ListMeta) {
	p.OnListMetaFunc(meta)
}

func (p Param) OnObject(unmarshaler proto.Unmarshaler) {
	p.OnObjectFunc(unmarshaler)
}

var protoEncodingPrefix = []byte{0x6b, 0x38, 0x73, 0x00}

func StreamList(ctx context.Context, client rest.Interface, resource string, namespace string, opts metav1.ListOptions, param ParamInterface) error {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	rc, err := client.Get().
		Namespace(namespace).
		Resource(resource).
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		SetHeader("Accept", runtime.ContentTypeProtobuf).
		Stream(ctx)
	if err != nil {
		return fmt.Errorf("client.Stream: %w", err)
	}
	defer rc.Close()

	prefix := make([]byte, len(protoEncodingPrefix))
	if _, err := io.ReadFull(rc, prefix); err != nil {
		return fmt.Errorf("rc.Read: %w", err)
	} else if !bytes.Equal(prefix, protoEncodingPrefix) {
		return errors.New("invalid protobuf encoding")
	}

	if err := (unknownStreamUnmarshaler{
		OnRaw: func(buffer *streamBuffer) error {
			if err := unmarshalListStream(buffer, param); err != nil {
				return fmt.Errorf("o.Unmarshal: %w", err)
			}
			return nil
		},
	}).Unmarshal(newStreamBuffer(rc, -1)); err != nil {
		return fmt.Errorf("unknown.Unmarshal: %w", err)
	}

	return nil
}
