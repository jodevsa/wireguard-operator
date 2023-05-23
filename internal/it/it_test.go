package it

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("wireguard controller", func() {
	It("wireguard is able to start", func() {
		wireguardYaml :=
			`apiVersion: vpn.example.com/v1alpha1
kind: Wireguard
metadata:
  name: vpn
spec:
  mtu: "1380"
  serviceType: "NodePort"`

		// kubectl apply -f
		output, err := KubectlApply(wireguardYaml, TestNamespace)
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(Equal("wireguard.vpn.example.com/vpn created"))

		wireguardPeerYaml :=
			`apiVersion: vpn.example.com/v1alpha1
kind: WireguardPeer
metadata:
  name: peer20
spec:
  wireguardRef: "vpn"

`
		// kubectl apply -f
		output, err = KubectlApply(wireguardPeerYaml, TestNamespace)
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(Equal("wireguardpeer.vpn.example.com/peer20 created"))

		WaitForWireguardToBeReady("vpn", TestNamespace)
		WaitForPeerToBeReady("peer20", TestNamespace)

		// TODO: connect to wg
	})
})
