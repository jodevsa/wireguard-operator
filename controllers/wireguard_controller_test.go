package controllers

import (
	"context"
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
		CronjobName      = "test-cronjob"
		CronjobNamespace = "default"
		JobName          = "test-job"
		Timeout          = time.Second * 10
		Interval         = time.Second * 1
	)

	Context("Wireguard with Mtu", func() {
		It("Should create succesfully", func() {
			By("By creating a new Wireguard")

			key := types.NamespacedName{
				Name:      "wireguard-with-mtu-1337",
				Namespace: "default",
			}
			created := &vpnv1alpha1.Wireguard{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: vpnv1alpha1.WireguardSpec{
					Mtu: "1337",
				},
			}
			labels := map[string]string{"app": "wireguard", "instance": key.Name}

			Expect(k8sClient.Create(context.Background(), created)).Should(Succeed())

			// service created
			externalHostName := "test-host-name"
			serviceName := key.Name + "-svc"
			serviceKey := types.NamespacedName{
				Namespace: key.Namespace,
				Name:      serviceName,
			}

			// match labels
			Eventually(func() map[string]string {
				svc := &corev1.Service{}
				k8sClient.Get(context.Background(), serviceKey, svc)
				return svc.Spec.Selector
			}, Timeout, Interval).Should(BeEquivalentTo(labels))

			// match service type
			Eventually(func() corev1.ServiceType {
				svc := &corev1.Service{}
				k8sClient.Get(context.Background(), serviceKey, svc)
				return svc.Spec.Type
			}, Timeout, Interval).Should(Equal(corev1.ServiceTypeLoadBalancer))

			// update service external hostname
			svc := &corev1.Service{}
			k8sClient.Get(context.Background(), serviceKey, svc)
			svc.Status.LoadBalancer.Ingress = []corev1.LoadBalancerIngress{{Hostname: externalHostName}}
			k8sClient.Status().Update(context.Background(), svc)

			// check that wireguard resource got the right hostname
			Eventually(func() vpnv1alpha1.WireguardStatus {
				wg := &vpnv1alpha1.Wireguard{}
				k8sClient.Get(context.Background(), key, wg)
				return wg.Status
			}, time.Second*50, Interval).Should(Equal(vpnv1alpha1.WireguardStatus{
				Hostname: externalHostName,
				Port:     "51820",
				Status:   "pending",
				Message:  "Fetching Wireguard status",
			}))

		})

	})

})
