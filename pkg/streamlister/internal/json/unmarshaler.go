package json

import (
	"encoding/json"
	"fmt"
	"io"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/ayanamist/k8s-utils/pkg/streamlister/internal/types"
)

func StreamUnmarshaler(r io.Reader, param types.ParamInterface) error {
	var typeMeta metav1.TypeMeta
	var apiVersionDecoded, kindDecoded bool

	dec := json.NewDecoder(r)
	dec.UseNumber()
	_, _ = dec.Token() // ignore `{`
Loop:
	for {
		k, err := dec.Token()
		if err != nil {
			return fmt.Errorf("dec.Token: %w", err)
		}
		switch k {
		case json.Delim('}'):
			break Loop
		case "apiVersion":
			if err := dec.Decode(&typeMeta.APIVersion); err != nil {
				return fmt.Errorf("decode apiVersion: %w", err)
			}
			apiVersionDecoded = true
			if kindDecoded {
				param.OnTypeMeta(&typeMeta)
			}
		case "kind":
			if err := dec.Decode(&typeMeta.Kind); err != nil {
				return fmt.Errorf("decode kind: %w", err)
			}
			kindDecoded = true
			if apiVersionDecoded {
				param.OnTypeMeta(&typeMeta)
			}
		case "metadata":
			listMeta := &metav1.ListMeta{}
			if err := dec.Decode(listMeta); err != nil {
				return fmt.Errorf("decode metadata: %w", err)
			}
			param.OnListMeta(listMeta)
		case "items":
			if t, err := dec.Token(); err != nil {
				return fmt.Errorf("decode items left bracket: %w", err)
			} else if t != json.Delim('[') {
				return fmt.Errorf("decode items but not array: %s", t)
			}
			for dec.More() {
				obj := param.ObjectFactory()
				if err := dec.Decode(obj); err != nil {
					return fmt.Errorf("decode item: %w", err)
				}
				param.OnObject(obj)
			}
			if t, err := dec.Token(); err != nil {
				return fmt.Errorf("decode items right bracket: %w", err)
			} else if t != json.Delim(']') {
				return fmt.Errorf("decode items but unexpected end: %s", t)
			}
		default:
			var ignored interface{}
			if err := dec.Decode(&ignored); err != nil {
				return fmt.Errorf("decode key=%s: %w", k, err)
			}
		}
	}
	if !kindDecoded || !apiVersionDecoded {
		param.OnTypeMeta(&typeMeta)
	}
	return nil
}
