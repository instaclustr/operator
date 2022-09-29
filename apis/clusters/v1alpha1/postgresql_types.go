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

type PostgreSQLDataCentre struct {
	DataCentre `json:",inline"`
	// PostgreSQL
	ClientEncryption      bool   `json:"clientEncryption,omitempty"`
	PostgresqlNodeCount   int32  `json:"postgresqlNodeCount"`
	ReplicationMode       string `json:"replicationMode,omitempty"`
	SynchronousModeStrict bool   `json:"synchronousModeStrict,omitempty"`
	// PGBouncer
	PoolMode string `json:"poolMode,omitempty"`
}

type PostgreSQLDataCentreStatus struct {
	DataCentreStatus `json:",inline"`
}

// PostgreSQLSpec defines the desired state of PostgreSQL
type PostgreSQLSpec struct {
	ClusterSpec `json:",inline"`
	PostgreSQLVersion string `json:"postgreSQLVersion"`
	PGBouncerVersion string `json:"pgBouncerVersion,omitempty"`
	DataCentres []*PostgreSQLDataCentre `json:"dataCentres,omitempty"`
}

// PostgreSQLStatus defines the observed state of PostgreSQL
type PostgreSQLStatus struct {
	ClusterStatus `json:",inline"`
	DataCentres   []*PostgreSQLDataCentreStatus `json:"dataCentres,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PostgreSQL is the Schema for the postgresqls API
type PostgreSQL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgreSQLSpec   `json:"spec,omitempty"`
	Status PostgreSQLStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PostgreSQLList contains a list of PostgreSQL
type PostgreSQLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgreSQL `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PostgreSQL{}, &PostgreSQLList{})
}
