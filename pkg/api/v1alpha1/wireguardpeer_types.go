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

	Address string `json:"address,omitempty"`
	// set to true to temporarily disable a peer
	Disabled   bool       `json:"disabled,omitempty"`
	Dns        string     `json:"dns,omitempty"`
	PrivateKey PrivateKey `json:"PrivateKeyRef,omitempty"`
	PublicKey  string     `json:"publicKey,omitempty"`
	// the name of the active wireguard instance
	//+kubebuilder:validation:Required
	//+kubebuilder:validation:MinLength=1
	WireguardRef string `json:"wireguardRef"`

	EgressNetworkPolicies EgressNetworkPolicies `json:"egressNetworkPolicies,omitempty"`
	DownloadSpeed         Speed                 `json:"downloadSpeed,omitempty"`
	UploadSpeed           Speed                 `json:"uploadSpeed,omitempty"`
}

type EgressNetworkPolicies []EgressNetworkPolicy

// +kubebuilder:validation:Enum=ACCEPT;REJECT;Accept;Reject
type EgressNetworkPolicyAction string

// +kubebuilder:validation:Enum=TCP;UDP
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
	Action   EgressNetworkPolicyAction   `json:"action,omitempty"`
	To       EgressNetworkPolicyTo       `json:"to,omitempty"`
	Protocol EgressNetworkPolicyProtocol `json:"protocol,omitempty"`
}

type EgressNetworkPolicyTo struct {
	Ip   string `json:"ip,omitempty"`
	Port int32  `json:"port,omitempty" protobuf:"varint,3,opt,name=port"`
}

// WireguardPeerStatus defines the observed state of WireguardPeer
type WireguardPeerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Config  string `json:"config,omitempty"`
	Status  string `json:"status,omitempty"`
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

	Spec   WireguardPeerSpec   `json:"spec,omitempty"`
	Status WireguardPeerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WireguardPeerList contains a list of WireguardPeer
type WireguardPeerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WireguardPeer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WireguardPeer{}, &WireguardPeerList{})
}
