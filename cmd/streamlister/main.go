package main

import (
	"context"
	"flag"
	"os"
	"runtime"

	"github.com/gogo/protobuf/proto"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	logMemStat("before list")

	podCh := make(chan *corev1.Pod, 5)

	go func() {
		defer close(podCh)
		if err := streamlister.StreamList(context.Background(), clientset.CoreV1().RESTClient(), "pods", "", metav1.ListOptions{ResourceVersion: "0"}, streamlister.Param{
			ObjectFactoryFunc: func() proto.Unmarshaler {
				return &corev1.Pod{}
			},
			OnListMetaFunc: func(listMeta *metav1.ListMeta) {
				logger.Infof("rv=%s", listMeta.ResourceVersion)
			},
			OnObjectFunc: func(obj proto.Unmarshaler) {
				pod := obj.(*corev1.Pod)
				podCh <- pod
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

	logMemStat("after request")
}

func logMemStat(message string) {
	var memStat runtime.MemStats
	runtime.ReadMemStats(&memStat)
	logrus.WithFields(logrus.Fields{
		"HeapAlloc": memStat.HeapAlloc,
		"HeapInuse": memStat.HeapInuse,
		"HeapSys":   memStat.HeapSys,
		"Sys":       memStat.Sys,
	}).Info(message)
}
