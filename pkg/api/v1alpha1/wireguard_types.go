/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"
)

const (
	Pending = "pending"
	Error   = "error"
	Ready   = "ready"
)

type WgStatusReport struct {
	// A string field that represents the current status of Wireguard. This could include values like ready, pending, or error.
	Status string `json:"status,omitempty"`
	// A string field that provides additional information about the status of Wireguard. This could include error messages or other information that helps to diagnose issues with the wg instance.
	Message string `json:"message,omitempty"`
}

// WireguardSpec defines the desired state of Wireguard
type WireguardSpec struct {
	// A string field that specifies the maximum transmission unit (MTU) size for Wireguard packets for all peers.
	Mtu string `json:"mtu,omitempty"`
	// A string field that specifies the address for the Wireguard VPN server. This is the public IP address or hostname that peers will use to connect to the VPN.
	Address string `json:"address,omitempty"`
	// A string field that specifies the DNS server(s) to be used by the peers.
	Dns string `json:"dns,omitempty"`
	// A field that specifies the type of Kubernetes service that should be used for the Wireguard VPN. This could be ClusterIP, NodePort or LoadBalancer, depending on the needs of the deployment.
	ServiceType corev1.ServiceType `json:"serviceType,omitempty"`
	// A field that specifies the value to use for a nodePort ServiceType
	NodePort int32 `json:"port,omitempty"`
	// A map of key value strings for service annotations
	ServiceAnnotations map[string]string `json:"serviceAnnotations,omitempty"`
	// A boolean field that specifies whether IP forwarding should be enabled on the Wireguard VPN pod at startup. This can be useful to enable if the peers are having problems with sending traffic to the internet.
	EnableIpForwardOnPodInit bool `json:"enableIpForwardOnPodInit,omitempty"`
	// A boolean field that specifies whether to use the userspace implementation of Wireguard instead of the kernel one.
	UseWgUserspaceImplementation bool `json:"useWgUserspaceImplementation,omitempty"`

	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	Agent        WireguardPodSpec  `json:"agent,omitempty"`
	Metric       WireguardPodSpec  `json:"metric,omitempty"`
}

// WireguardPodSpec defines spec for respective containers created for Wireguard
type WireguardPodSpec struct {
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
}

// WireguardStatus defines the observed state of Wireguard
type WireguardStatus struct {
	// A string field that specifies the address for the Wireguard VPN server that is currently being used.
	Address         string `json:"address,omitempty"`
	Dns             string `json:"dns,omitempty"`
	DnsSearchDomain string `json:"dnssearchdomain,omitempty"`
	PublicKey       string `json:"publickey,omitempty"`
	// A string field that specifies the port for the Wireguard VPN server that is currently being used.
	Port string `json:"port,omitempty"`
	// A string field that represents the current status of Wireguard. This could include values like ready, pending, or error.
	Status string `json:"status,omitempty"`
	// A string field that provides additional information about the status of Wireguard. This could include error messages or other information that helps to diagnose issues with the wg instance.
	Message          string     `json:"message,omitempty"`
	Resources        []Resource `json:"resources,omitempty"`
	UniqueIdentifier string     `json:"UniqueIdentifier,omitempty"`
}
type Resource struct {
	Name   string `json:"name,omitempty"`
	Type   string `json:"type,omitempty"`
	Status string `json:"status,omitempty"`
}

type DeploymentResource struct {
	Name   string `json:"name,omitempty"`
	Status string `json:"status,omitempty"`
	Image  string `json:"image,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Wireguard is the Schema for the wireguards API
type Wireguard struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WireguardSpec   `json:"spec,omitempty"`
	Status WireguardStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WireguardList contains a list of Wireguard
type WireguardList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Wireguard `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Wireguard{}, &WireguardList{})
}
