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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GCPVPCPeeringSpec defines the desired state of GCPVPCPeering
type GCPVPCPeeringSpec struct {
	VPCPeeringSpec     `json:",inline"`
	PeerVPCNetworkName string `json:"peerVpcNetworkName"`
	PeerProjectID      string `json:"peerProjectId"`
}

// GCPVPCPeeringStatus defines the observed state of GCPVPCPeering
type GCPVPCPeeringStatus struct {
	VPCPeeringStatus `json:",inline"`
	Name             string `json:"name"`
	FailureReason    string `json:"failureReason"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GCPVPCPeering is the Schema for the gcpvpcpeerings API
type GCPVPCPeering struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GCPVPCPeeringSpec   `json:"spec,omitempty"`
	Status GCPVPCPeeringStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GCPVPCPeeringList contains a list of GCPVPCPeering
type GCPVPCPeeringList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GCPVPCPeering `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GCPVPCPeering{}, &GCPVPCPeeringList{})
}
