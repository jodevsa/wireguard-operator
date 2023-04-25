package iptables

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/jodevsa/wireguard-operator/pkg/agent"
	"github.com/jodevsa/wireguard-operator/pkg/api/v1alpha1"
	"os"
	"os/exec"
	"strings"
)

func ApplyRules(rules string) error {
	file, err := os.CreateTemp("/tmp", "iptables-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(file.Name())

	err = os.WriteFile(file.Name(), []byte(rules), 0640)

	if err != nil {
		return err
	}

	bashCommand := fmt.Sprintf("iptables-restore < %s", file.Name())
	cmd := exec.Command("bash", "-c", bashCommand)

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

type Iptables struct {
	Logger logr.Logger
}

func (it *Iptables) Sync(state agent.State) error {
	it.Logger.Info("syncing network policies")
	wgHostName := state.Server.Status.Address
	dns := state.Server.Status.Dns
	peers := state.Peers

	cfg := GenerateIptableRulesFromPeers(wgHostName, dns, peers)

	err := ApplyRules(cfg)

	if err != nil {
		return err
	}

	return nil
}

func GenerateIptableRulesFromNetworkPolicies(policies v1alpha1.EgressNetworkPolicies, peerIp string, kubeDnsIp string, wgServerIp string) string {
	var rules []string

	// add a comment
	rules = append(rules, fmt.Sprintf("# start of rules for peer %s", peerIp))

	peerChain := strings.ReplaceAll(peerIp, ".", "-")

	// create chain for peer
	rules = append(rules, fmt.Sprintf(":%s - [0:0]", peerChain))
	// associate peer chain to FORWARD chain
	rules = append(rules, fmt.Sprintf("-A FORWARD -s %s -j %s", peerIp, peerChain))

	// allow peer to ping (ICMP) wireguard server for debugging purposes
	rules = append(rules, fmt.Sprintf("-A %s -d %s -p icmp -j ACCEPT", peerChain, wgServerIp))
	// allow peer to communicate with itself
	rules = append(rules, fmt.Sprintf("-A %s -d %s -j ACCEPT", peerChain, peerIp))
	// allow peer to communicate with kube-dns
	rules = append(rules, fmt.Sprintf("-A %s -d %s -p UDP --dport 53 -j ACCEPT", peerChain, kubeDnsIp))

	for _, policy := range policies {
		for _, rule := range EgressNetworkPolicyToIpTableRules(policy, peerChain) {
			rules = append(rules, rule)
		}
	}

	// if policies are defined impose an implicit deny all
	if len(policies) != 0 {
		rules = append(rules, fmt.Sprintf("-A %s -j REJECT --reject-with icmp-port-unreachable", peerChain))
	}

	// add a comment
	rules = append(rules, fmt.Sprintf("# end of rules for peer %s", peerIp))

	return strings.Join(rules, "\n")
}

func GenerateIptableRulesFromPeers(wgHostName string, dns string, peers []v1alpha1.WireguardPeer) string {
	var rules []string

	var natTableRules = fmt.Sprintf(`
*nat
:PREROUTING ACCEPT [0:0]
:INPUT ACCEPT [0:0]
:OUTPUT ACCEPT [0:0]
:POSTROUTING ACCEPT [0:0]
-A POSTROUTING -s 10.8.0.0/24 -o eth0 -j MASQUERADE
COMMIT`)

	for _, peer := range peers {

		//tc(peer.Spec.DownloadSpeed, peer.Spec.UploadSpeed)
		rules = append(rules, GenerateIptableRulesFromNetworkPolicies(peer.Spec.EgressNetworkPolicies, peer.Spec.Address, dns, wgHostName))
	}

	var filterTableRules = fmt.Sprintf(`
*filter
:INPUT ACCEPT [0:0]
:FORWARD ACCEPT [0:0]
:OUTPUT ACCEPT [0:0]
%s
COMMIT
`, strings.Join(rules, "\n"))

	return fmt.Sprintf("%s\n%s", natTableRules, filterTableRules)
}

func EgressNetworkPolicyToIpTableRules(policy v1alpha1.EgressNetworkPolicy, peerChain string) []string {

	var rules []string

	if policy.Protocol == "" && policy.To.Port != 0 {
		policy.Protocol = "TCP"
		rules = append(rules, EgressNetworkPolicyToIpTableRules(policy, peerChain)[0])
		policy.Protocol = "UDP"
		rules = append(rules, EgressNetworkPolicyToIpTableRules(policy, peerChain)[0])
		return rules
	}

	// customer rules
	var rulePeerChain = "-A " + peerChain
	var ruleAction = string("-j " + v1alpha1.EgressNetworkPolicyActionDeny)
	var ruleProtocol = ""
	var ruleDestIp = ""
	var ruleDestPort = ""

	if policy.To.Ip != "" {
		ruleDestIp = "-d " + policy.To.Ip
	}

	if policy.Protocol != "" {
		ruleProtocol = "-p " + strings.ToUpper(string(policy.Protocol))
	}

	if policy.To.Port != 0 {
		ruleDestPort = "--dport " + fmt.Sprint(policy.To.Port)
	}

	if policy.Action != "" {
		ruleAction = "-j " + strings.ToUpper(string(policy.Action))
	}

	var options = []string{rulePeerChain, ruleDestIp, ruleProtocol, ruleDestPort, ruleAction}
	var filteredOptions []string
	for _, option := range options {
		if len(option) != 0 {
			filteredOptions = append(filteredOptions, option)
		}
	}
	rules = append(rules, strings.Join(filteredOptions, " "))

	return rules

}
