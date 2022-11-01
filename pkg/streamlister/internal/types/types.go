package types

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type ParamInterface interface {
	ObjectFactory() runtime.Object
	OnListMeta(*metav1.ListMeta)
	OnTypeMeta(*metav1.TypeMeta)
	OnObject(runtime.Object)
}

type ParamFuncs struct {
	ObjectFactoryFunc func() runtime.Object
	OnListMetaFunc    func(*metav1.ListMeta)
	OnTypeMetaFunc    func(*metav1.TypeMeta)
	OnObjectFunc      func(runtime.Object)
}

func (p ParamFuncs) ObjectFactory() runtime.Object {
	return p.ObjectFactoryFunc()
}

func (p ParamFuncs) OnListMeta(meta *metav1.ListMeta) {
	onListMetaFunc := p.OnListMetaFunc
	if onListMetaFunc != nil {
		onListMetaFunc(meta)
	}
}

func (p ParamFuncs) OnTypeMeta(meta *metav1.TypeMeta) {
	onTypeMetaFunc := p.OnTypeMetaFunc
	if onTypeMetaFunc != nil {
		onTypeMetaFunc(meta)
	}
}

func (p ParamFuncs) OnObject(o runtime.Object) {
	p.OnObjectFunc(o)
}
