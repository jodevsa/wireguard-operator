package iptables

import (
	"github.com/jodevsa/wireguard-operator/pkg/api/v1alpha1"
	"testing"
)

// test helpers

func TestIptableRules(t *testing.T) {
	tests := []struct {
		name                 string
		peerIp               string
		kubeDnsIp            string
		wgServerIp           string
		networkPolicies      v1alpha1.EgressNetworkPolicies
		expectedIptableRules string
	}{
		{
			name:       "EgressNetworkPolicy with destination IP address filter",
			peerIp:     "192.168.1.115",
			kubeDnsIp:  "69.96.1.42",
			wgServerIp: "192.168.1.1",
			networkPolicies: v1alpha1.EgressNetworkPolicies{
				v1alpha1.EgressNetworkPolicy{
					Action: v1alpha1.EgressNetworkPolicyActionAccept,
					To:     v1alpha1.EgressNetworkPolicyTo{Ip: "8.8.8.8"}},
			},
			expectedIptableRules: `# start of rules for peer 192.168.1.115
:192-168-1-115 - [0:0]
-A FORWARD -s 192.168.1.115 -j 192-168-1-115
-A 192-168-1-115 -d 192.168.1.1 -p icmp -j ACCEPT
-A 192-168-1-115 -d 192.168.1.115 -j ACCEPT
-A 192-168-1-115 -d 69.96.1.42 -p UDP --dport 53 -j ACCEPT
-A 192-168-1-115 -d 8.8.8.8 -j ACCEPT
-A 192-168-1-115 -j REJECT --reject-with icmp-port-unreachable
# end of rules for peer 192.168.1.115`,
		},
		{
			name:       "Able to filter egress by UDP",
			peerIp:     "10.8.0.9",
			kubeDnsIp:  "100.64.0.10",
			wgServerIp: "10.8.0.1",
			networkPolicies: v1alpha1.EgressNetworkPolicies{
				v1alpha1.EgressNetworkPolicy{
					Action:   v1alpha1.EgressNetworkPolicyActionAccept,
					Protocol: "UDP",
					To:       v1alpha1.EgressNetworkPolicyTo{}},
			},
			expectedIptableRules: `# start of rules for peer 10.8.0.9
:10-8-0-9 - [0:0]
-A FORWARD -s 10.8.0.9 -j 10-8-0-9
-A 10-8-0-9 -d 10.8.0.1 -p icmp -j ACCEPT
-A 10-8-0-9 -d 10.8.0.9 -j ACCEPT
-A 10-8-0-9 -d 100.64.0.10 -p UDP --dport 53 -j ACCEPT
-A 10-8-0-9 -p UDP -j ACCEPT
-A 10-8-0-9 -j REJECT --reject-with icmp-port-unreachable
# end of rules for peer 10.8.0.9`,
		},
		{
			name:            "Empty networkPolicies",
			peerIp:          "10.8.0.9",
			kubeDnsIp:       "100.64.0.10",
			wgServerIp:      "10.8.0.1",
			networkPolicies: v1alpha1.EgressNetworkPolicies{},
			expectedIptableRules: `# start of rules for peer 10.8.0.9
:10-8-0-9 - [0:0]
-A FORWARD -s 10.8.0.9 -j 10-8-0-9
-A 10-8-0-9 -d 10.8.0.1 -p icmp -j ACCEPT
-A 10-8-0-9 -d 10.8.0.9 -j ACCEPT
-A 10-8-0-9 -d 100.64.0.10 -p UDP --dport 53 -j ACCEPT
# end of rules for peer 10.8.0.9`,
		},
		{
			name:            "networkPolicies with 1 empty networkPolicy",
			peerIp:          "10.8.0.11",
			kubeDnsIp:       "100.64.0.21",
			wgServerIp:      "10.7.0.1",
			networkPolicies: v1alpha1.EgressNetworkPolicies{v1alpha1.EgressNetworkPolicy{}},
			expectedIptableRules: `# start of rules for peer 10.8.0.11
:10-8-0-11 - [0:0]
-A FORWARD -s 10.8.0.11 -j 10-8-0-11
-A 10-8-0-11 -d 10.7.0.1 -p icmp -j ACCEPT
-A 10-8-0-11 -d 10.8.0.11 -j ACCEPT
-A 10-8-0-11 -d 100.64.0.21 -p UDP --dport 53 -j ACCEPT
-A 10-8-0-11 -j Reject
-A 10-8-0-11 -j REJECT --reject-with icmp-port-unreachable
# end of rules for peer 10.8.0.11`,
		},
		{
			name:       "EgressNetworkPolicy with destination port Allowed",
			peerIp:     "10.8.0.9",
			kubeDnsIp:  "100.64.0.10",
			wgServerIp: "10.8.0.1",
			networkPolicies: v1alpha1.EgressNetworkPolicies{v1alpha1.EgressNetworkPolicy{
				Protocol: v1alpha1.EgressNetworkPolicyProtocolTCP,
				Action:   v1alpha1.EgressNetworkPolicyActionAccept,
				To:       v1alpha1.EgressNetworkPolicyTo{Port: 8080},
			}},
			expectedIptableRules: `# start of rules for peer 10.8.0.9
:10-8-0-9 - [0:0]
-A FORWARD -s 10.8.0.9 -j 10-8-0-9
-A 10-8-0-9 -d 10.8.0.1 -p icmp -j ACCEPT
-A 10-8-0-9 -d 10.8.0.9 -j ACCEPT
-A 10-8-0-9 -d 100.64.0.10 -p UDP --dport 53 -j ACCEPT
-A 10-8-0-9 -p TCP --dport 8080 -j ACCEPT
-A 10-8-0-9 -j REJECT --reject-with icmp-port-unreachable
# end of rules for peer 10.8.0.9`,
		},
	}

	for _, test := range tests {

		t.Run(test.name, func(t *testing.T) {

			rules := GenerateIptableRulesFromNetworkPolicies(test.networkPolicies, test.peerIp, test.kubeDnsIp, test.wgServerIp)
			if rules != test.expectedIptableRules {
				t.Errorf("got %s, want %s", rules, test.expectedIptableRules)
			}
		})
	}
}
