module github.com/snapp-cab/gslb-controller

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/m-yosefpor/utils v0.0.0-20210703235507-8b7d90bca5db
	github.com/onsi/ginkgo v1.14.1
	github.com/onsi/gomega v1.10.2
	github.com/snapp-incubator/consul-gslb-driver v1.2.0
	google.golang.org/grpc v1.38.0
	k8s.io/api v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/klog/v2 v2.9.0
	sigs.k8s.io/controller-runtime v0.8.3
)
