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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PrivateKey struct {
	SecretKeyRef corev1.SecretKeySelector `json:"secretKeyRef"`
}

type Status struct {
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WireguardPeerSpec defines the desired state of WireguardPeer
type WireguardPeerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The address of the peer.
	Address string `json:"address,omitempty"`
	// The AllowedIPs of the peer.
	AllowedIPs string `json:"allowedIPs,omitempty"`
	// Set to true to temporarily disable the peer.
	Disabled bool `json:"disabled,omitempty"`
	// The DNS configuration for the peer.
	Dns string `json:"dns,omitempty"`
	// The private key of the peer
	PrivateKey PrivateKey `json:"privateKeyRef,omitempty"`
	// The key used by the peer to authenticate with the wg server.
	PublicKey string `json:"publicKey,omitempty"`
	// The name of the Wireguard instance in k8s that the peer belongs to. The wg instance should be in the same namespace as the peer.
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=1
	WireguardRef string `json:"wireguardRef"`
	// Egress network policies for the peer.
	EgressNetworkPolicies EgressNetworkPolicies `json:"egressNetworkPolicies,omitempty"`
	DownloadSpeed         Speed                 `json:"downloadSpeed,omitempty"`
	UploadSpeed           Speed                 `json:"uploadSpeed,omitempty"`
}

type EgressNetworkPolicies []EgressNetworkPolicy

// +kubebuilder:validation:Enum=ACCEPT;REJECT;Accept;Reject
type EgressNetworkPolicyAction string

// +kubebuilder:validation:Enum=TCP;UDP;ICMP
type EgressNetworkPolicyProtocol string

const (
	EgressNetworkPolicyActionAccept EgressNetworkPolicyAction = "Accept"
	EgressNetworkPolicyActionDeny   EgressNetworkPolicyAction = "Reject"
)

const (
	EgressNetworkPolicyProtocolTCP EgressNetworkPolicyProtocol = "TCP"
	EgressNetworkPolicyProtocolUDP EgressNetworkPolicyProtocol = "UDP"
)

type EgressNetworkPolicy struct {
	// Specifies the action to take when outgoing traffic from a Wireguard peer matches the policy. This could be 'Accept' or 'Reject'.
	Action EgressNetworkPolicyAction `json:"action,omitempty"`
	// A struct that specifies the destination address and port for the traffic. This could include IP addresses or hostnames, as well as specific port numbers or port ranges.
	To EgressNetworkPolicyTo `json:"to,omitempty"`
	// Specifies the protocol to match for this policy. This could be TCP, UDP, or ICMP.
	Protocol EgressNetworkPolicyProtocol `json:"protocol,omitempty"`
}

type EgressNetworkPolicyTo struct {
	// A string field that specifies the destination IP address for traffic that matches the policy.
	Ip string `json:"ip,omitempty"`
	// An integer field that specifies the destination port number for traffic that matches the policy.
	Port int32 `json:"port,omitempty" protobuf:"varint,3,opt,name=port"`
}

// WireguardPeerStatus defines the observed state of WireguardPeer
type WireguardPeerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// A string field that contains the current configuration for the Wireguard peer.
	Config string `json:"config,omitempty"`
	// A string field that represents the current status of the Wireguard peer. This could include values like ready, pending, or error.
	Status string `json:"status,omitempty"`
	// A string field that provides additional information about the status of the Wireguard peer. This could include error messages or other information that helps to diagnose issues with the peer.
	Message string `json:"message,omitempty"`
}

type Speed struct {
	Value int `json:"config,omitempty"`

	// +kubebuilder:validation:Enum=mbps;kbps
	Unit string `json:"unit,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// WireguardPeer is the Schema for the wireguardpeers API
type WireguardPeer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// The desired state of the peer.
	Spec WireguardPeerSpec `json:"spec,omitempty"`
	// A field that defines the observed state of the Wireguard peer. This includes fields like the current configuration and status of the peer.
	Status WireguardPeerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WireguardPeerList contains a list of WireguardPeer
type WireguardPeerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WireguardPeer `json:"items"`
}

func (in *WireguardPeerList) Error() string {
	panic("implement me")
}

func init() {
	SchemeBuilder.Register(&WireguardPeer{}, &WireguardPeerList{})
}
