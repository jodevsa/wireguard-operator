module github.com/jodevsa/wireguard-operator

go 1.16

require (
	github.com/emicklei/go-restful v2.9.5+incompatible // indirect
	github.com/fsnotify/fsnotify v1.6.0
	github.com/go-logr/logr v1.2.4
	github.com/go-logr/stdr v1.2.2
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/korylprince/ipnetgen v1.0.1
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/ginkgo/v2 v2.9.5 // indirect
	github.com/onsi/gomega v1.27.6
	github.com/vishvananda/netlink v1.1.0
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20230429144221-925a1e7659e6
	k8s.io/api v0.26.1
	k8s.io/apimachinery v0.26.1
	k8s.io/client-go v0.26.1
	sigs.k8s.io/controller-runtime v0.14.6
	sigs.k8s.io/kind v0.19.0
)
