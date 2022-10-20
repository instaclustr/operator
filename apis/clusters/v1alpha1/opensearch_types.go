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

// OpenSearchSpec defines the desired state of OpenSearch
type OpenSearchSpec struct {
	Cluster               `json:",inline"`
	DataCentres           []*OpenSearchDataCentre `json:"dataCentres"`
	ConcurrentResizes     int                     `json:"concurrentResizes,omitempty"`
	NotifySupportContacts bool                    `json:"notifySupportContacts,omitempty"`
	ClusterConfigurations map[string]string       `json:"clusterConfigurations,omitempty"`
	Description           string                  `json:"description,omitempty"`
	PrivateLink           *PrivateLink            `json:"privateLink,omitempty"`
}

type OpenSearchDataCentre struct {
	DataCentre                   `json:",inline"`
	DedicatedMasterNodes         bool   `json:"dedicatedMasterNodes,omitempty"`
	MasterNodeSize               string `json:"masterNodeSize,omitempty"`
	OpenSearchDashboardsNodeSize string `json:"openSearchDashboardsNodeSize,omitempty"`
	IndexManagementPlugin        bool   `json:"indexManagementPlugin,omitempty"`
}

// OpenSearchStatus defines the observed state of OpenSearch
type OpenSearchStatus struct {
	ClusterStatus `json:",inline"`
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
