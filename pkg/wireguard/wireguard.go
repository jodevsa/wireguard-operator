package wireguard

import (
	"github.com/jodevsa/wireguard-operator/pkg/agent"
	"github.com/vishvananda/netlink"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"net"
	"syscall"
)

const MTU = 1420

func syncRoute(_ agent.State, iface string) error {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return err
	}

	routes, err := netlink.RouteList(link, syscall.AF_INET)
	if err != nil {
		return err
	}

	for _, route := range routes {
		if route.LinkIndex == link.Attrs().Index {
			return nil
		}
	}
	route := netlink.Route{
		LinkIndex: link.Attrs().Index,
		Dst:       &getIP("10.8.0.0/24")[0],
		Gw:        net.ParseIP("10.8.0.1"),
	}

	err = netlink.RouteAdd(&route)
	if err != nil {
		return err
	}

	return nil
}

func syncAddress(_ agent.State, iface string) error {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		return err
	}

	addresses, err := netlink.AddrList(link, syscall.AF_INET)
	if err != nil {
		return nil
	}

	if len(addresses) != 0 {
		return nil
	}
	err = netlink.AddrAdd(link, &netlink.Addr{
		IPNet: &net.IPNet{IP: net.ParseIP("10.8.0.1")},
	})

	if err := netlink.LinkSetUp(link); err != nil {
		return err
	}
	return nil
}

func SyncLink(_ agent.State, iface string) error {
	link, err := netlink.LinkByName(iface)
	if err != nil {
		if _, ok := err.(netlink.LinkNotFoundError); !ok {
			return err
		}
	}

	if _, ok := err.(netlink.LinkNotFoundError); ok {
		// link not created
		wgLink := &netlink.GenericLink{
			LinkAttrs: netlink.LinkAttrs{
				Name: iface,
				MTU:  MTU,
			},
			LinkType: "wireguard",
		}
		// create
		if err := netlink.LinkAdd(wgLink); err != nil {
			return err
		}

		link, err = netlink.LinkByName(iface)

		if err != nil {
			return err
		}
		if err := netlink.LinkSetUp(link); err != nil {
			return err
		}
	}

	link, err = netlink.LinkByName(iface)
	if err != nil {
		if _, ok := err.(netlink.LinkNotFoundError); !ok {
			return err
		}
	}

	addresses, err := netlink.AddrList(link, syscall.AF_INET)
	if err != nil {
		return nil
	}

	if len(addresses) != 0 {
		return nil
	}
	err = netlink.AddrAdd(link, &netlink.Addr{
		IPNet: &getIP("10.8.0.1/32")[0],
	})

	if err := netlink.LinkSetUp(link); err != nil {
		return err
	}
	return nil
}

func syncWireguard(state agent.State, iface string, listenPort int) error {
	c, _ := wgctrl.New()
	cfg, err := CreateWireguardConfiguration(state, listenPort)
	if err != nil {
		return err
	}
	err = c.ConfigureDevice(iface, cfg)
	if err != nil {
		return err
	}

	return nil
}

func Sync(state agent.State, iface string, listenPort int) error {
	// create wg0 link
	err := SyncLink(state, iface)
	if err != nil {
		return err
	}

	// create route
	err = syncRoute(state, iface)

	if err != nil {
		return err
	}

	// set wg0 address to 10.8.0.1/32
	err = syncAddress(state, iface)
	if err != nil {
		return err
	}

	// sync wg configuration
	err = syncWireguard(state, iface, listenPort)
	if err != nil {
		return err
	}

	return nil
}

func getIP(ip string) []net.IPNet {
	_, ipnet, _ := net.ParseCIDR(ip)

	return []net.IPNet{*ipnet}
}

func CreateWireguardConfiguration(state agent.State, listenPort int) (wgtypes.Config, error) {
	cfg := wgtypes.Config{}

	key, err := wgtypes.ParseKey(state.ServerPrivateKey)
	if err != nil {
		return wgtypes.Config{}, err
	}
	cfg.PrivateKey = &key

	cfg.ReplacePeers = true
	cfg.ListenPort = &listenPort

	var peers []wgtypes.PeerConfig

	for _, peer := range state.Peers {
		if peer.Spec.Disabled == true {
			continue
		}
		if peer.Spec.PublicKey == "" {
			continue
		}

		if peer.Spec.Address == "" {
			continue
		}

		peerCfg := wgtypes.PeerConfig{AllowedIPs: getIP(peer.Spec.Address + "/32")}

		key, err := wgtypes.ParseKey(peer.Spec.PublicKey)
		if err != nil {
			return wgtypes.Config{}, err
		}
		peerCfg.PublicKey = key
		peerCfg.ReplaceAllowedIPs = true

		peers = append(peers, peerCfg)
	}

	cfg.Peers = peers

	return cfg, nil
}
