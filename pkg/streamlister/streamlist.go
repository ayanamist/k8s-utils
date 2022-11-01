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
	"k8s.io/apimachinery/pkg/util/naming"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/utils/trace"

	"github.com/ayanamist/k8s-utils/pkg/streamlister/internal/json"
	"github.com/ayanamist/k8s-utils/pkg/streamlister/internal/protobuf"
	"github.com/ayanamist/k8s-utils/pkg/streamlister/internal/types"
)

type ParamInterface = types.ParamInterface

type ParamFuncs = types.ParamFuncs

type streamListOptions struct {
	traceThreshold time.Duration
	parameterCodec runtime.ParameterCodec
}

func createDefaultOptions() *streamListOptions {
	return &streamListOptions{
		traceThreshold: 10 * time.Second,
		parameterCodec: scheme.ParameterCodec,
	}
}

type OptionFunc func(options *streamListOptions)

func WithTraceThreshold(threshold time.Duration) OptionFunc {
	return func(options *streamListOptions) {
		options.traceThreshold = threshold
	}
}

func WithParameterCodec(codec runtime.ParameterCodec) OptionFunc {
	return func(options *streamListOptions) {
		options.parameterCodec = codec
	}
}

func StreamList(ctx context.Context, client rest.Interface, resource string, namespace string, listOptions metav1.ListOptions, param ParamInterface, opts ...OptionFunc) error {
	slo := createDefaultOptions()
	for _, opt := range opts {
		opt(slo)
	}

	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}

	initTrace := trace.New("StreamList", trace.Field{Key: "name", Value: naming.GetNameFromCallsite()})
	defer initTrace.LogIfLong(slo.traceThreshold)

	const acceptContentType = runtime.ContentTypeProtobuf + "," + runtime.ContentTypeJSON
	rc, err := client.Get().
		Namespace(namespace).
		Resource(resource).
		VersionedParams(&listOptions, slo.parameterCodec).
		Timeout(timeout).
		SetHeader("Accept", acceptContentType).
		Stream(ctx)

	initTrace.Step("APIServer responded")

	if err != nil {
		return fmt.Errorf("client.Stream: %w", err)
	}
	defer rc.Close()

	var encoding string
	defer initTrace.Step("Transfer and unmarshal finished: " + encoding)

	prefix := make([]byte, 1)
	if _, err := io.ReadFull(rc, prefix); err != nil {
		return fmt.Errorf("rc.Read 1st byte: %w", err)
	}
	if firstByte := prefix[0]; firstByte == protobuf.EncodingPrefix[0] {
		prefix := make([]byte, len(protobuf.EncodingPrefix)-1)
		if _, err := io.ReadFull(rc, prefix); err != nil {
			return fmt.Errorf("rc.Read: %w", err)
		} else if !bytes.Equal(prefix, protobuf.EncodingPrefix[1:]) {
			return errors.New("invalid protobuf encoding")
		}

		encoding = "protobuf"

		o := param.ObjectFactory()
		if _, ok := o.(proto.Unmarshaler); !ok {
			return fmt.Errorf("object %T returned by ObjectFactory does not implement proto.Unmarshaler", o)
		}

		if err := (protobuf.UnknownStreamUnmarshaler{
			OnRaw: func(buffer *protobuf.StreamBuffer) error {
				if err := protobuf.UnmarshalListStream(buffer, param); err != nil {
					return fmt.Errorf("protobuf.UnmarshalListStream: %w", err)
				}
				return nil
			},
			OnTypeMeta: param.OnTypeMeta,
		}).Unmarshal(protobuf.NewStreamBuffer(rc, -1)); err != nil {
			return fmt.Errorf("protobuf.UnmarshalUnknown: %w", err)
		}
	} else if firstByte == '{' {
		encoding = "json"
		if err := json.StreamUnmarshaler(io.MultiReader(bytes.NewReader(prefix), rc), param); err != nil {
			return fmt.Errorf("json.Unmarshal: %w", err)
		}
	}

	return nil
}
