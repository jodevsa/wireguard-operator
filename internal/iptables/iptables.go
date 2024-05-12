package iptables

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/go-logr/logr"
	"github.com/jodevsa/wireguard-operator/pkg/agent"
	"github.com/jodevsa/wireguard-operator/pkg/api/v1alpha1"
)

func ApplyRules(rules string) error {
	cmd := exec.Command("iptables-restore")
	cmd.Stdin = strings.NewReader(rules)
	return cmd.Run()
}

type Iptables struct {
	Logger logr.Logger
}

func (it *Iptables) Sync(state agent.State) error {
	it.Logger.Info("syncing network policies")
	wgHostName := state.Server.Status.Address
	dns := state.Server.Status.Dns
	peers := state.Peers

	cfg := GenerateIPTableRulesFromPeers(wgHostName, dns, peers)

	err := ApplyRules(cfg)

	if err != nil {
		return err
	}

	return nil
}

func GenerateIPTableRulesFromNetworkPolicies(policies v1alpha1.EgressNetworkPolicies, peerIp string, kubeDnsIp string, wgServerIp string) string {
	peerChain := strings.ReplaceAll(peerIp, ".", "-")

	rules := []string{
		// add a comment
		fmt.Sprintf("# start of rules for peer %s", peerIp),

		// create chain for peer
		fmt.Sprintf(":%s - [0:0]", peerChain),

		// associate peer chain to FORWARD chain
		fmt.Sprintf("-A FORWARD -s %s -j %s", peerIp, peerChain),

		// allow peer to ping (ICMP) wireguard server for debugging purposes
		fmt.Sprintf("-A %s -d %s -p icmp -j ACCEPT", peerChain, wgServerIp),

		// allow peer to communicate with itself
		fmt.Sprintf("-A %s -d %s -j ACCEPT", peerChain, peerIp),

		// allow peer to communicate with kube-dns
		fmt.Sprintf("-A %s -d %s -p UDP --dport 53 -j ACCEPT", peerChain, kubeDnsIp),
	}

	for _, policy := range policies {
		rules = append(rules, EgressNetworkPolicyToIPTableRules(policy, peerChain)...)
	}

	// if policies are defined impose an implicit deny all
	if len(policies) != 0 {
		rules = append(rules, fmt.Sprintf("-A %s -j REJECT --reject-with icmp-port-unreachable", peerChain))
	}

	// add a comment
	rules = append(rules, fmt.Sprintf("# end of rules for peer %s", peerIp))

	return strings.Join(rules, "\n")
}

func GenerateIPTableRulesFromPeers(wgHostName string, dns string, peers []v1alpha1.WireguardPeer) string {
	const natTableRules = `*nat
:PREROUTING ACCEPT [0:0]
:INPUT ACCEPT [0:0]
:OUTPUT ACCEPT [0:0]
:POSTROUTING ACCEPT [0:0]
-A POSTROUTING -s 10.8.0.0/24 -o eth0 -j MASQUERADE
COMMIT`

	var rules []string
	for _, peer := range peers {
		rules = append(rules, GenerateIPTableRulesFromNetworkPolicies(peer.Spec.EgressNetworkPolicies, peer.Spec.Address, dns, wgHostName))
	}

	filterTableRules := fmt.Sprintf(`*filter
:INPUT ACCEPT [0:0]
:FORWARD ACCEPT [0:0]
:OUTPUT ACCEPT [0:0]
%s
COMMIT`, strings.Join(rules, "\n"))

	return fmt.Sprintf("%s\n%s", natTableRules, filterTableRules)
}

func EgressNetworkPolicyToIPTableRules(policy v1alpha1.EgressNetworkPolicy, peerChain string) []string {
	if policy.Protocol == "" && policy.To.Port != 0 {
		tcpPolicy := policy
		tcpPolicy.Protocol = "TCP"

		udpPolicy := policy
		udpPolicy.Protocol = "UDP"

		return []string{
			EgressNetworkPolicyToIPTableRule(tcpPolicy, peerChain),
			EgressNetworkPolicyToIPTableRule(udpPolicy, peerChain),
		}
	}
	return []string{EgressNetworkPolicyToIPTableRule(policy, peerChain)}
}

func EgressNetworkPolicyToIPTableRule(policy v1alpha1.EgressNetworkPolicy, peerChain string) string {
	opts := []string{fmt.Sprintf("-A %s", peerChain)}

	if policy.To.Ip != "" {
		opts = append(opts, fmt.Sprintf("-d %s", policy.To.Ip))
	}

	if policy.Protocol != "" {
		opts = append(opts, fmt.Sprintf("-p %s", strings.ToUpper(string(policy.Protocol))))
	}

	if policy.To.Port != 0 {
		opts = append(opts, fmt.Sprintf("--dport %d", policy.To.Port))
	}

	action := v1alpha1.EgressNetworkPolicyActionDeny
	if policy.Action != "" {
		action = policy.Action
	}

	opts = append(opts, fmt.Sprintf("-j %s", strings.ToUpper(string(action))))

	return strings.Join(opts, " ")
}
