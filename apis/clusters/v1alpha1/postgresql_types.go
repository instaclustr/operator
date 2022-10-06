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

type PgDataCentre struct {
	DataCentre `json:",inline"`
	// PostgreSQL options
	ClientEncryption      bool   `json:"clientEncryption,omitempty"`
	PostgresqlNodeCount   int32  `json:"postgresqlNodeCount"`
	ReplicationMode       string `json:"replicationMode,omitempty"`
	SynchronousModeStrict bool   `json:"synchronousModeStrict,omitempty"`
	// PGBouncer options
	PoolMode string `json:"poolMode,omitempty"`
}

// PgSpec defines the desired state of PostgreSQL
type PgSpec struct {
	Cluster `json:",inline"`

	// running as dependency for another instance.
	// need to be tested before decide what to do with this field
	BundledUseOnlyCluster bool            `json:"bundledUseOnlyCluster,omitempty"`
	OIDCProvider          string          `json:"oidsProvider,omitempty"`
	FirewallRules         []*FirewallRule `json:"firewallRules,omitempty"`
	PGBouncerVersion      string          `json:"pgBouncerVersion,omitempty"`
	DataCentres           []*PgDataCentre `json:"dataCentres"`
}

// PgStatus defines the observed state of PostgreSQL
type PgStatus struct {
	ClusterStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PostgreSQL is the Schema for the postgresqls API
type PostgreSQL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PgSpec   `json:"spec,omitempty"`
	Status PgStatus `json:"status,omitempty"`
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
