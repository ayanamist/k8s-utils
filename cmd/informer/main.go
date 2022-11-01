package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	stdruntime "runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/metadata"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	"github.com/ayanamist/k8s-utils/pkg/streamlister"
)

type PodPtrList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []*corev1.Pod
}

func (m *PodPtrList) DeepCopyObject() runtime.Object {
	c := &PodPtrList{}
	c.TypeMeta = m.TypeMeta
	m.ListMeta.DeepCopyInto(&c.ListMeta)
	c.Items = make([]*corev1.Pod, len(m.Items))
	for i, p := range m.Items {
		c.Items[i] = p.DeepCopy()
	}
	return c
}

func listPod(ctx context.Context, client rest.Interface, namespace string, options metav1.ListOptions) (runtime.Object, error) {
	podList := &PodPtrList{}
	err := streamlister.StreamList(ctx, client, "pods", namespace, options, streamlister.ParamFuncs{
		ObjectFactoryFunc: func() runtime.Object {
			return &corev1.Pod{}
		},
		OnListMetaFunc: func(meta *metav1.ListMeta) {
			podList.ListMeta = *meta
		},
		OnObjectFunc: func(o runtime.Object) {
			podList.Items = append(podList.Items, o.(*corev1.Pod))
		},
		OnTypeMetaFunc: func(meta *metav1.TypeMeta) {
			podList.TypeMeta = *meta
		},
	})
	return podList, err
}

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	logger := logrus.StandardLogger()

	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		logger.WithError(err).Fatal("clientcmd.BuildConfigFromFlags failed")
	}
	config = metadata.ConfigFor(config)

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.WithError(err).Fatal("kubernetes.NewForConfig failed")
	}

	ctx := context.TODO()

	logMemStat("before informer started")

	informerFactory := startInformerFactory(ctx, client)

	logMemStat("after informer started")

	podInformerStore := informerFactory.Core().V1().Pods().Informer().GetStore()
	keys := podInformerStore.ListKeys()
	logger.Infof("found %d", len(keys))
}

func startInformerFactory(ctx context.Context, client *kubernetes.Clientset) informers.SharedInformerFactory {
	informerFactory := informers.NewSharedInformerFactory(client, time.Hour)

	informerFactory.InformerFor(&corev1.Pod{}, func(client kubernetes.Interface, resyncDur time.Duration) cache.SharedIndexInformer {
		const namespace = ""
		return cache.NewSharedIndexInformer(
			&cache.ListWatch{
				ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
					opt, _ := json.Marshal(options)
					logrus.Infof("list options=%s", opt)
					return listPod(ctx, client.CoreV1().RESTClient(), namespace, options)
					//return client.CoreV1().Pods(namespace).List(ctx, options)
				},
				WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
					return client.CoreV1().Pods(namespace).Watch(ctx, options)
				},
			},
			&corev1.Pod{},
			resyncDur,
			cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc},
		)
	})

	informerFactory.Start(ctx.Done())
	informerFactory.WaitForCacheSync(ctx.Done())

	return informerFactory
}

func logMemStat(message string) {
	b, _ := ioutil.ReadFile("/proc/self/smaps_rollup")
	var rss string
	for scanner := bufio.NewScanner(bytes.NewReader(b)); scanner.Scan(); {
		if line := scanner.Text(); strings.HasPrefix(line, "Rss:") {
			rss = strings.TrimSpace(line[4:])
		}
	}
	var memStat stdruntime.MemStats
	stdruntime.ReadMemStats(&memStat)
	logrus.WithFields(logrus.Fields{
		"HeapInuse": memStat.HeapInuse,
		"RSS":       rss,
	}).Info(message)
}
