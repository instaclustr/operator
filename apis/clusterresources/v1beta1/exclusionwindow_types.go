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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ExclusionWindowSpec defines the desired state of ExclusionWindow
type ExclusionWindowSpec struct {
	ClusterID string `json:"clusterId"`
	DayOfWeek string `json:"dayOfWeek"`
	//+kubebuilder:validation:Minimum:=0
	//+kubebuilder:validation:Maximum:=23
	StartHour int32 `json:"startHour"`

	//kubebuilder:validation:Minimum:=1
	//+kubebuilder:validation:Maximum:=40
	DurationInHours int32 `json:"durationInHours"`
}

// ExclusionWindowStatus defines the observed state of ExclusionWindow
type ExclusionWindowStatus struct {
	ID string `json:"id"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
//+kubebuilder:printcolumn:name="ID",type="string",JSONPath=".status.id"

// ExclusionWindow is the Schema for the exclusionwindows API
type ExclusionWindow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExclusionWindowSpec   `json:"spec,omitempty"`
	Status ExclusionWindowStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ExclusionWindowList contains a list of ExclusionWindow
type ExclusionWindowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExclusionWindow `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ExclusionWindow{}, &ExclusionWindowList{})
}

func (r *ExclusionWindow) NewPatch() client.Patch {
	old := r.DeepCopy()
	return client.MergeFrom(old)
}
