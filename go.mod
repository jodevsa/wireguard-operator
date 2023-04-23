module github.com/jodevsa/wireguard-operator

go 1.16

require (
	github.com/fsnotify/fsnotify v1.6.0
	github.com/korylprince/ipnetgen v1.0.1
	github.com/mdlayher/netlink v1.7.1 // indirect
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/vishvananda/netlink v1.1.0
	golang.zx2c4.com/wireguard/wgctrl v0.0.0-20211215182854-7a385b3431de
	k8s.io/api v0.22.1
	k8s.io/apimachinery v0.22.1
	k8s.io/client-go v0.22.1
	sigs.k8s.io/controller-runtime v0.10.0
)
