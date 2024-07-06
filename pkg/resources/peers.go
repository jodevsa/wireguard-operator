package resources

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/jodevsa/wireguard-operator/pkg/api/v1alpha1"
	"github.com/korylprince/ipnetgen"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Peers struct {
	Wireguard *v1alpha1.Wireguard
	Logger    logr.Logger
	Client    client.Client
}

func (p Peers) Type() string {
	return "peers"
}

func (p Peers) Converged(ctx context.Context) (bool, error) {
	return true, nil
}

func (p Peers) Name() string {
	return "peers"
}

func (p Peers) Create(ctx context.Context) error {
	return p.Update(ctx)
}

func (p Peers) NeedsUpdate(ctx context.Context) (bool, error) {
	peers, err := p.getWireguardPeers(ctx)

	for _, peer := range peers.Items {
		if peer.Spec.Address == "" {
			return true, nil
		}

		cfg := p.getPeerConfig(peer)
		if peer.Status.Config != cfg || peer.Status.Status != v1alpha1.Ready {
			return true, nil

		}
	}
	if err != nil {
		return true, nil
	}

	return false, nil
}
func (p Peers) getPeerConfig(peer v1alpha1.WireguardPeer) string {
	cfg := fmt.Sprintf(`
echo "
[Interface]
PrivateKey = $(kubectl get secret %s-peer --template={{.data.privateKey}} -n %s | base64 -d)
Address = %s
DNS = %s`, peer.Name, peer.Namespace, peer.Spec.Address, p.constructDnsConfig())

	if p.Wireguard.Spec.Mtu != "" {
		cfg = cfg + "\nMTU = " + p.Wireguard.Spec.Mtu
	}

	return cfg + fmt.Sprintf(`

[Peer]
PublicKey = %s
AllowedIPs = 0.0.0.0/0
Endpoint = %s:%s"`, p.Wireguard.Status.PublicKey, p.Wireguard.Status.Address, p.Wireguard.Status.Port)

}

func (p Peers) constructDnsConfig() string {
	dnsConfiguration := p.Wireguard.Status.Dns

	if p.Wireguard.Status.DnsSearchDomain != "" {
		dnsConfiguration = p.Wireguard.Status.Dns + ", " + p.Wireguard.Status.DnsSearchDomain
	}

	return dnsConfiguration
}

func (p Peers) Update(ctx context.Context) error {
	peers, err := p.getWireguardPeers(ctx)

	usedIps := p.getUsedIps(peers)

	if err != nil {
		return err
	}

	for _, peer := range peers.Items {
		if peer.Spec.Address == "" {
			ip, err := getAvaialbleIp("10.8.0.0/24", usedIps)

			if err != nil {
				return err
			}

			peer.Spec.Address = ip

			if err := p.Client.Update(ctx, &peer); err != nil {
				return err
			}

			usedIps = append(usedIps, ip)
		}

		newConfig := p.getPeerConfig(peer)

		if peer.Status.Config != newConfig || peer.Status.Status != v1alpha1.Ready {
			peer.Status.Config = newConfig
			peer.Status.Status = v1alpha1.Ready
			peer.Status.Message = "Peer configured"
			if err := p.Client.Status().Update(ctx, &peer); err != nil {
				return err
			}
		}
	}

	return nil
}

func getAvaialbleIp(cidr string, usedIps []string) (string, error) {
	gen, err := ipnetgen.New(cidr)
	if err != nil {
		return "", err
	}
	for ip := gen.Next(); ip != nil; ip = gen.Next() {
		used := false
		for _, usedIp := range usedIps {
			if ip.String() == usedIp {
				used = true
				break
			}
		}
		if !used {
			return ip.String(), nil
		}
	}

	return "", fmt.Errorf("no available ip found in %s", cidr)
}

func (r *Peers) getUsedIps(peers *v1alpha1.WireguardPeerList) []string {
	usedIps := []string{"10.8.0.0", "10.8.0.1"}
	for _, p := range peers.Items {
		usedIps = append(usedIps, p.Spec.Address)

	}

	return usedIps
}

func (r *Peers) getWireguardPeers(ctx context.Context) (*v1alpha1.WireguardPeerList, error) {
	peers := &v1alpha1.WireguardPeerList{}
	if err := r.Client.List(ctx, peers, client.InNamespace(r.Wireguard.Namespace)); err != nil {
		return nil, err
	}

	relatedPeers := &v1alpha1.WireguardPeerList{}

	for _, peer := range peers.Items {
		if peer.Spec.WireguardRef == r.Wireguard.Name {
			relatedPeers.Items = append(relatedPeers.Items, peer)
		}
	}

	return relatedPeers, nil
}
