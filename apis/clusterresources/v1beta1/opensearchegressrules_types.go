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

type OpenSearchEgressRulesSpec struct {
	ClusterID           string `json:"clusterId"`
	OpenSearchBindingID string `json:"openSearchBindingId"`
	Source              string `json:"source"`
	Type                string `json:"type,omitempty"`
}

type OpenSearchEgressRulesStatus struct {
	ID string `json:"id,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OpenSearchEgressRules is the Schema for the opensearchegressrules API
type OpenSearchEgressRules struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenSearchEgressRulesSpec   `json:"spec,omitempty"`
	Status OpenSearchEgressRulesStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OpenSearchEgressRulesList contains a list of OpenSearchEgressRules
type OpenSearchEgressRulesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenSearchEgressRules `json:"items"`
}

func (er *OpenSearchEgressRules) NewPatch() client.Patch {
	old := er.DeepCopy()
	return client.MergeFrom(old)
}

func init() {
	SchemeBuilder.Register(&OpenSearchEgressRules{}, &OpenSearchEgressRulesList{})
}
