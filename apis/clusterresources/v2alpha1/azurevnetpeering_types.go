/*
Copyright 2022.

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

package v2alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AzureVNetPeeringSpec defines the desired state of AzureVNetPeering
type AzureVNetPeeringSpec struct {
	VPCPeeringSpec         `json:",inline"`
	PeerResourceGroup      string `json:"peerResourceGroup"`
	PeerSubscriptionID     string `json:"peerSubscriptionId"`
	PeerADObjectID         string `json:"peerAdObjectId,omitempty"`
	PeerVirtualNetworkName string `json:"peerVirtualNetworkName"`
}

// AzureVNetPeeringStatus defines the observed state of AzureVNetPeering
type AzureVNetPeeringStatus struct {
	VPCPeeringStatus `json:",inline"`
	Name             string `json:"name"`
	FailureReason    string `json:"failureReason"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AzureVNetPeering is the Schema for the azurevnetpeerings API
type AzureVNetPeering struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureVNetPeeringSpec   `json:"spec,omitempty"`
	Status AzureVNetPeeringStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AzureVNetPeeringList contains a list of AzureVNetPeering
type AzureVNetPeeringList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureVNetPeering `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AzureVNetPeering{}, &AzureVNetPeeringList{})
}
