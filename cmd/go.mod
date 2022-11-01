module github.com/ayanamist/k8s-utils/cmd

go 1.15

require (
	github.com/ayanamist/k8s-utils v0.0.0
	github.com/sirupsen/logrus v1.9.0
	k8s.io/api v0.20.15
	k8s.io/apimachinery v0.20.15
	k8s.io/client-go v0.20.15
	k8s.io/klog/v2 v2.4.0
)

replace github.com/ayanamist/k8s-utils => ../
