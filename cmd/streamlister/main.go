package main

import (
	"context"
	"flag"
	"os"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/ayanamist/k8s-utils/pkg/streamlister"
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	logger := logrus.StandardLogger()

	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		logger.WithError(err).Fatal("clientcmd.BuildConfigFromFlags failed")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.WithError(err).Fatal("kubernetes.NewForConfig failed")
	}

	podCh := make(chan *corev1.Pod, 5)

	go func() {
		defer close(podCh)
		if err := streamlister.StreamList(context.Background(), clientset.CoreV1().RESTClient(), "pods", "", metav1.ListOptions{ResourceVersion: "0"}, streamlister.ParamFuncs{
			ObjectFactoryFunc: func() runtime.Object {
				return &corev1.Pod{}
			},
			OnListMetaFunc: func(listMeta *metav1.ListMeta) {
				logger.Infof("rv=%s", listMeta.ResourceVersion)
			},
			OnObjectFunc: func(obj runtime.Object) {
				pod := obj.(*corev1.Pod)
				podCh <- pod
			},
			OnTypeMetaFunc: func(meta *metav1.TypeMeta) {
				logger.Infof("typeMeta: %+v", meta)
			},
		}); err != nil {
			logger.WithError(err).Fatal("streamlister.StreamList failed")
		}
	}()

	var i int
	for range podCh {
		i++
	}

	logger.Infof("podList count=%d", i)
}
