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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type CassandraAWSSettings struct {
	EBSEncryptionKey       string `json:"ebsEncryptionKey,omitempty"`
	CustomVirtualNetworkID string `json:"customVirtualNetworkId,omitempty"`
}

type CassandraGCPSettings struct {
	CustomVirtualNetworkID string `json:"customVirtualNetworkId,omitempty"`
}

type CassandraAzureSettings struct {
	ResourceGroup string `json:"resourceGroup,omitempty"`
}

type Spark struct {
	Version string `json:"version"`
}

type CassandraDataCentre struct {
	ContinuousBackup               bool                      `json:"continuousBackup"`
	ReplicationFactor              int                       `json:"replicationFactor"`
	AWSSettings                    []*CassandraAWSSettings   `json:"awsSettings,omitempty"`
	NumberOfNodes                  int                       `json:"numberOfNodes"`
	PrivateIPBroadcastForDiscovery bool                      `json:"privateIpBroadcastForDiscovery"`
	Network                        string                    `json:"network,omitempty"`
	Tags                           []*Tag                    `json:"tags,omitempty"`
	GCPSettings                    []*CassandraGCPSettings   `json:"gcpSettings,omitempty"`
	ClientToClusterEncryption      bool                      `json:"clientToClusterEncryption"`
	NodeSize                       string                    `json:"nodeSize"`
	CloudProvider                  string                    `json:"cloudProvider"`
	AzureSettings                  []*CassandraAzureSettings `json:"azureSettings,omitempty"`
	Name                           string                    `json:"name"`
	Region                         string                    `json:"region"`
	ProviderAccountName            string                    `json:"providerAccountName,omitempty"`
}

type CassandraNode struct {
	NodeRoles      []string `json:"nodeRoles,omitempty"`
	Rack           string   `json:"rack,omitempty"`
	NodeSize       string   `json:"nodeSize,omitempty"`
	PrivateAddress string   `json:"privateAddress,omitempty"`
	PublicAddress  string   `json:"publicAddress,omitempty"`
	NodeID         string   `json:"nodeID,omitempty"`
	NodeStatus     string   `json:"nodeStatus,omitempty"`
}

// CassandraSpec defines the desired state of Cassandra
type CassandraSpec struct {
	PCIComplianceMode     bool                   `json:"pciComplianceMode"`
	TwoFactorDelete       []*TwoFactorDelete     `json:"twoFactorDelete,omitempty"`
	DataCentres           []*CassandraDataCentre `json:"dataCentres"`
	PrivateNetworkCluster bool                   `json:"privateNetworkCluster"`
	CassandraVersion      string                 `json:"cassandraVersion"`
	LuceneEnabled         bool                   `json:"luceneEnabled"`
	PasswordAndUserAuth   bool                   `json:"passwordAndUserAuth"`
	Name                  string                 `json:"name"`
	SLATier               string                 `json:"slaTier"`
}

type CassandraDCStatus struct {
	ID             string  `json:"id,omitempty"`
	Status         string  `json:"status,omitempty"`
	CassandraNodes []*Node `json:"nodes,omitempty"`
}

// CassandraStatus defines the observed state of Cassandra
type CassandraStatus struct {
	ClusterID         string               `json:"id,omitempty"`
	ClusterStatus     string               `json:"status,omitempty"`
	CassandraDCStatus []*CassandraDCStatus `json:"dataCentres,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Cassandra is the Schema for the cassandras API
type Cassandra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CassandraSpec   `json:"spec,omitempty"`
	Status CassandraStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CassandraList contains a list of Cassandra
type CassandraList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cassandra `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cassandra{}, &CassandraList{})
}
