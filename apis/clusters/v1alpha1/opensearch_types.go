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

type OpenSearchBundle struct {
	Bundle  `json:",inline"`
	Options OpenSearchBundleOptions `json:"options,omitempty"`
}

type OpenSearchBundleOptions struct {
	DedicatedMasterNodes         bool   `json:"dedicatedMasterNodes,omitempty"`
	MasterNodeSize               string `json:"masterNodeSize,omitempty"`
	OpenSearchDashboardsNodeSize string `json:"openSearchDashboardsNodeSize,omitempty"`
	IndexManagementPlugin        bool   `json:"indexManagementPlugin,omitempty"`
}

type OpenSearchDataCentre struct {
	GenericDataCentre `json:",inline"`
	Bundles           []OpenSearchBundle `json:"bundles,omitempty"`
}

type OpenSearchDataCentreStatus struct {
	DataCentreStatus `json:",inline"`
	Nodes            []Node `json:"nodes,omitempty"`
}

// OpenSearchSpec defines the desired state of OpenSearch
type OpenSearchSpec struct {
	GenericCluster `json:",inline"`
	Bundles        []OpenSearchBundle     `json:"bundles"`
	DataCentres    []OpenSearchDataCentre `json:"dataCentres,omitempty"`
}

// OpenSearchStatus defines the observed state of OpenSearch
type OpenSearchStatus struct {
	ClusterStatus `json:",inline"`
	DataCentres   []DataCentreStatus `json:"dataCentres,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// OpenSearch is the Schema for the opensearches API
type OpenSearch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenSearchSpec   `json:"spec,omitempty"`
	Status OpenSearchStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// OpenSearchList contains a list of OpenSearch
type OpenSearchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OpenSearch `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OpenSearch{}, &OpenSearchList{})
}
