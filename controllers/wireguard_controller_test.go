package controllers

import (
	"context"
	"fmt"
	"time"

	vpnv1alpha1 "github.com/jodevsa/wireguard-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("wireguard controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		wgName       = "vpn"
		wgNamespace  = "default"
		Timeout      = time.Second * 2
		Interval     = time.Second * 1
		dnsServiceIp = "10.0.0.42"
	)

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
				Status:  "ready",
				Message: "Peer configured",
			}))

		})

	})

})
