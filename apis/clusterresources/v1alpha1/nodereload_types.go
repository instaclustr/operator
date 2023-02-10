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
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NodeReloadSpec defines the desired state of NodeReload
type NodeReloadSpec struct {
	Nodes []*Node `json:"nodes"`
}

// NodeReloadStatus defines the observed state of NodeReload
type NodeReloadStatus struct {
	NodeInProgress         Node         `json:"nodeInProgress,omitempty"`
	CurrentOperationStatus []*Operation `json:"currentOperationStatus,omitempty"`
}

type Node struct {
	Bundle string `json:"bundle"`
	NodeID string `json:"nodeID"`
}

type Operation struct {
	TimeCreated  int64  `json:"timeCreated"`
	TimeModified int64  `json:"timeModified"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NodeReload is the Schema for the nodereloads API
type NodeReload struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeReloadSpec   `json:"spec,omitempty"`
	Status NodeReloadStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodeReloadList contains a list of NodeReload
type NodeReloadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeReload `json:"items"`
}

func (nr *NodeReload) NewPatch() client.Patch {
	old := nr.DeepCopy()
	return client.MergeFrom(old)
}

func init() {
	SchemeBuilder.Register(&NodeReload{}, &NodeReloadList{})
}
