module github.com/jodevsa/wireguard-operator

go 1.16

require (
	github.com/emicklei/go-restful v2.9.5+incompatible // indirect
	github.com/fsnotify/fsnotify v1.6.0
	github.com/go-logr/logr v1.2.3
	github.com/go-logr/stdr v1.2.2
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/korylprince/ipnetgen v1.0.1
	github.com/mdlayher/netlink v1.7.1 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.24.1
	github.com/vishvananda/netlink v1.1.0
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20230215201556-9c5414ab4bde
	k8s.io/api v0.26.1
	k8s.io/apimachinery v0.26.1
	k8s.io/client-go v0.26.1
	sigs.k8s.io/controller-runtime v0.14.6
)
