package controllers

import (
	"context"
	"fmt"
	"time"

	vpnv1alpha1 "github.com/jodevsa/wireguard-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("wireguard controller", func() {

	Context("Egress Network Policy", func() {
		tests := []struct {
			name                 string
			peerIp               string
			kubeDnsIp            string
			wgServerIp           string
			networkPolicies      vpnv1alpha1.EgressNetworkPolicies
			expectedIptableRules string
		}{
			{
				name:       "EgressNetworkPolicy with destination IP address filter",
				peerIp:     "192.168.1.115",
				kubeDnsIp:  "69.96.1.42",
				wgServerIp: "192.168.1.1",
				networkPolicies: vpnv1alpha1.EgressNetworkPolicies{
					vpnv1alpha1.EgressNetworkPolicy{
						Action: vpnv1alpha1.EgressNetworkPolicyActionAccept,
						To:     vpnv1alpha1.EgressNetworkPolicyTo{Ip: "8.8.8.8"}},
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
				networkPolicies: vpnv1alpha1.EgressNetworkPolicies{
					vpnv1alpha1.EgressNetworkPolicy{
						Action:   vpnv1alpha1.EgressNetworkPolicyActionAccept,
						Protocol: "UDP",
						To:       vpnv1alpha1.EgressNetworkPolicyTo{}},
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
				networkPolicies: vpnv1alpha1.EgressNetworkPolicies{},
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
				networkPolicies: vpnv1alpha1.EgressNetworkPolicies{vpnv1alpha1.EgressNetworkPolicy{}},
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
				networkPolicies: vpnv1alpha1.EgressNetworkPolicies{vpnv1alpha1.EgressNetworkPolicy{
					Protocol: vpnv1alpha1.EgressNetworkPolicyProtocolTCP,
					Action:   vpnv1alpha1.EgressNetworkPolicyActionAccept,
					To:       vpnv1alpha1.EgressNetworkPolicyTo{Port: 8080},
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

			It(test.name, func() {
				Expect(EgressNetworkPoliciestoIptableRules(test.networkPolicies, test.peerIp, test.kubeDnsIp, test.wgServerIp)).Should(Equal(test.expectedIptableRules))
			})
		}
	})

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		wgTestImage  = "test-image"
		wgName       = "vpn"
		wgNamespace  = "default"
		Timeout      = time.Second * 2
		Interval     = time.Second * 1
		dnsServiceIp = "10.0.0.42"
	)

	viper.Set("WIREGUARD_IMAGE", wgTestImage)

	Context("Wireguard", func() {

		It("Should create succesfully", func() {
			// create kube-dns service
			dnsService := &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kube-dns",
					Namespace: "kube-system",
				},
				Spec: corev1.ServiceSpec{
					ClusterIP: dnsServiceIp,
					Ports:     []corev1.ServicePort{{Name: "dns", Protocol: corev1.ProtocolUDP, Port: 53}},
				},
			}

			Expect(k8sClient.Create(context.Background(), dnsService)).Should(Succeed())

			wgKey := types.NamespacedName{
				Name:      wgName,
				Namespace: wgNamespace,
			}
			created := &vpnv1alpha1.Wireguard{
				ObjectMeta: metav1.ObjectMeta{
					Name:      wgKey.Name,
					Namespace: wgKey.Namespace,
				},
			}
			expectedLabels := map[string]string{"app": "wireguard", "instance": wgKey.Name}

			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			// service created
			expectedExternalHostName := "test-host-name"
			serviceName := wgKey.Name + "-svc"
			serviceKey := types.NamespacedName{
				Namespace: wgKey.Namespace,
				Name:      serviceName,
			}

			// match labels
			Eventually(func() map[string]string {
				svc := &corev1.Service{}
				k8sClient.Get(context.Background(), serviceKey, svc)
				return svc.Spec.Selector
			}, Timeout, Interval).Should(BeEquivalentTo(expectedLabels))

			// match service type
			Eventually(func() corev1.ServiceType {
				svc := &corev1.Service{}
				k8sClient.Get(context.Background(), serviceKey, svc)
				return svc.Spec.Type
			}, Timeout, Interval).Should(Equal(corev1.ServiceTypeLoadBalancer))

			Eventually(func() vpnv1alpha1.WireguardStatus {
				wg := &vpnv1alpha1.Wireguard{}
				k8sClient.Get(context.Background(), wgKey, wg)
				return wg.Status
			}, Timeout, Interval).Should(Equal(vpnv1alpha1.WireguardStatus{
				Hostname: "",
				Port:     "",
				Status:   "pending",
				Message:  "Waiting for service to be ready",
			}))

			// update service external hostname
			svc := &corev1.Service{}
			k8sClient.Get(context.Background(), serviceKey, svc)
			svc.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{{Hostname: expectedExternalHostName}}
			Expect(k8sClient.Status().Update(context.Background(), svc)).Should(Succeed())

			// check that wireguard resource got the right status after the service is ready
			wg := &vpnv1alpha1.Wireguard{}
			Eventually(func() vpnv1alpha1.WireguardStatus {
				Expect(k8sClient.Get(context.Background(), wgKey, wg)).Should(Succeed())
				return wg.Status
			}, Timeout, Interval).Should(Equal(vpnv1alpha1.WireguardStatus{
				Hostname: expectedExternalHostName,
				Port:     "51820",
				Status:   "ready",
				Message:  "VPN is active!",
			}))

			Eventually(func() string {
				deploymentKey := types.NamespacedName{
					Name:      wgName + "-dep",
					Namespace: wgNamespace,
				}
				deployment := &appsv1.Deployment{}
				Expect(k8sClient.Get(context.Background(), deploymentKey, deployment)).Should(Succeed())
				Expect(len(deployment.Spec.Template.Spec.Containers)).Should(Equal(2))
				Expect(deployment.Spec.Template.Spec.Containers[0].Image).Should(Equal(deployment.Spec.Template.Spec.Containers[1].Image))
				return deployment.Spec.Template.Spec.Containers[0].Image
			}, Timeout, Interval).Should(Equal(wgTestImage))

			// create peer
			peerKey := types.NamespacedName{
				Name:      wgKey.Name + "peer",
				Namespace: wgKey.Namespace,
			}
			peer := &vpnv1alpha1.WireguardPeer{
				ObjectMeta: metav1.ObjectMeta{
					Name:      peerKey.Name,
					Namespace: peerKey.Namespace,
				},
				Spec: vpnv1alpha1.WireguardPeerSpec{
					WireguardRef: wgKey.Name,
				},
			}
			Expect(k8sClient.Create(context.Background(), peer)).Should(Succeed())

			//get peer secret
			wgSecretKeyName := types.NamespacedName{
				Name:      wgKey.Name,
				Namespace: wgKey.Namespace,
			}
			wgSecret := &corev1.Secret{}
			Eventually(func() error {
				return k8sClient.Get(context.Background(), wgSecretKeyName, wgSecret)
			}, Timeout, Interval).Should(Succeed())
			wgPublicKey := string(wgSecret.Data["publicKey"])

			Eventually(func() string {
				Expect(k8sClient.Get(context.Background(), peerKey, peer)).Should(Succeed())
				return peer.Spec.Address
			}, Timeout, Interval).Should(Equal("10.8.0.2"))

			Eventually(func() vpnv1alpha1.WireguardPeerStatus {
				Expect(k8sClient.Get(context.Background(), peerKey, peer)).Should(Succeed())
				return peer.Status
			}, Timeout, Interval).Should(Equal(vpnv1alpha1.WireguardPeerStatus{
				Config: fmt.Sprintf(`
echo "
[Interface]
PrivateKey = $(kubectl get secret %s-peer --template={{.data.privateKey}} -n default | base64 -d)
Address = %s
DNS = %s, %s.svc.cluster.local

[Peer]
PublicKey = %s
AllowedIPs = 0.0.0.0/0
Endpoint = %s:%s"`, peerKey.Name, peer.Spec.Address, dnsServiceIp, peer.Namespace, wgPublicKey, expectedExternalHostName, wg.Status.Port),
				Status:       "ready",
				Message:      "Peer configured",
				IptableRules: "# start of rules for peer 10.8.0.2\n:10-8-0-2 - [0:0]\n-A FORWARD -s 10.8.0.2 -j 10-8-0-2\n-A 10-8-0-2 -d test-host-name -p icmp -j ACCEPT\n-A 10-8-0-2 -d 10.8.0.2 -j ACCEPT\n-A 10-8-0-2 -d 10.0.0.42, default.svc.cluster.local -p UDP --dport 53 -j ACCEPT\n# end of rules for peer 10.8.0.2",
			}))

		})

	})

})
