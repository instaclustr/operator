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

type CadenceBundle struct {
	Bundle  `json:",inline"`
	Options CadenceBundleOptions `json:"options,omitempty"`
}

type CadenceBundleOptions struct {
	UseAdvancedVisibility   bool   `json:"useAdvancedVisibility,omitempty"`
	TargetCassandraCDCID    string `json:"targetCassandraCdcId"`
	TargetCassandraVPCType  string `json:"targetCassandraVpcType,omitempty"`
	TargetKafkaCDCID        string `json:"targetKafkaCdcId,omitempty"`
	TargetKafkaVPCType      string `json:"targetKafkaVpcType,omitempty"`
	TargetOpenSearchCDCID   string `json:"targetOpenSearchCdcId"`
	TargetOpenSearchVPCType string `json:"targetOpenSearchVpcType"`
	EnableArchival          bool   `json:"enableArchival,omitempty"`
	ArchivalS3URI           string `json:"archivalS3Uri,omitempty"`
	ArchivalS3Region        string `json:"archivalS3Region,omitempty"`
	AWSAccessKeySecretName  string `json:"awsAccessKeySecretName,omitempty"`
}

type CadenceDataCentre struct {
	GenericDataCentre `json:",inline"`
	Bundles           []*CadenceBundle `json:"bundles,omitempty"`
}

type CadenceDataCentreStatus struct {
	DataCentreStatus `json:",inline"`
	Nodes            []*Node `json:"nodes,omitempty"`
}

// CadenceSpec defines the desired state of Cadence
type CadenceSpec struct {
	GenericCluster   `json:",inline"`
	Bundles          []*CadenceBundle `json:"bundles"`
	PackagedSolution bool             `json:"packaged_solution,omitempty"`
}

// CadenceSpec defines the observed state of Cadence
type CadenceStatus struct {
	ClusterStatus `json:",inline"`
	DataCentres   []*DataCentreStatus `json:"dataCentres,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Cadence is the Schema for the cadences API
type Cadence struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CadenceSpec   `json:"spec,omitempty"`
	Status CadenceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CadenceList contains a list of Cadence
type CadenceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cadence `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cadence{}, &CadenceList{})
}
