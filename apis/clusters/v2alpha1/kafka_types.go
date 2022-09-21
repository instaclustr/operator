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

type SchemaRegistry struct {
	Version string `json:"version"`
}

type RestProxy struct {
	IntegrateRestProxyWithSchemaRegistry bool   `json:"integrateRestProxyWithSchemaRegistry"`
	UseLocalSchemaRegistry               bool   `json:"useLocalSchemaRegistry,omitempty"`
	SchemaRegistryServerURL              string `json:"schemaRegistryServerUrl,omitempty"`
	SchemaRegistryUsername               string `json:"schemaRegistryUsername,omitempty"`
	SchemaRegistryPassword               string `json:"schemaRegistryPassword,omitempty"`
	Version                              string `json:"version"`
}

type DedicatedZookeeper struct {
	ZookeeperNodeSize  string `json:"zookeeperNodeSize"`
	ZookeeperNodeCount int32  `json:"zookeeperNodeCount"`
}

type KafkaDataCentre struct {
	// A logical Name for the data centre within a cluster. These names must be unique in the cluster.
	Name string `json:"name"`

	// The private network address block for the Data Centre specified using CIDR address notation.
	// The Network must have a prefix length between /12 and /22 and must be part of a private address space.
	Network string `json:"network"`

	// NodeSize is a size of the nodes provisioned in the Data Centre.
	NodeSize string `json:"nodeSize"`

	// Total number of Kafka brokers in the Data Centre. Must be a multiple of defaultReplicationFactor.
	NumberOfNodes int32 `json:"numberOfNodes"`

	// AWS specific settings for the Data Centre. Cannot be provided with GCPSettings and AzureSettings.
	AWSSettings []*AWSSettings `json:"awsSettings,omitempty"`

	// GCPSettings specific settings for the Data Centre. Cannot be provided with AWSSettings and AzureSettings.
	GCPSettings []*GCPSetting `json:"gcpSettings,omitempty"`

	// AzureSettings specific settings for the Data Centre. Cannot be provided with GCPSettings and AWSSettings.
	AzureSettings []*AzureSettings `json:"azureSettings,omitempty"`

	// List of tags to apply to the Data Centre. Tags are metadata labels which allow you to identify,
	// categorize and filter clusters. This can be useful for grouping together clusters into applications,
	// environments, or any category that you require.
	Tags []*Tag `json:"tags"`

	// Enum: "AWS_VPC" "GCP" "AZURE" "AZURE_AZ"
	// CloudProvider is name of the cloud provider service in which the Data Centre will be provisioned.
	CloudProvider string `json:"cloudProvider"`

	// Region of the Data Centre.
	Region string `json:"region"`

	// For customers running in their own account. Your provider account can be found on the Create Cluster page on
	// the Instaclustr Console, or the "Provider Account" property on any existing cluster. For customers provisioning on
	// Instaclustr's cloud provider accounts, this property may be omitted.
	ProviderAccountName string `json:"providerAccountName,omitempty"`
}

type AWSSettings struct {
	// ID of a KMS encryption key to encrypt data on nodes.
	// KMS encryption key must be set in Cluster Resources through the Instaclustr Console before
	// provisioning an encrypted Data Centre.
	EBSEncryptionKey string `json:"ebsEncryptionKey,omitempty"`

	// VPC ID into which the Data Centre will be provisioned.
	// The Data Centre's network allocation must match the IPv4 CIDR block of the specified VPC.
	CustomVirtualNetworkID string `json:"customVirtualNetworkId,omitempty"`
}

type GCPSetting struct {
	// Network name or a relative Network or Subnetwork URI e.g. projects/my-project/regions/us-central1/subnetworks/my-subnet.
	// The Data Centre's network allocation must match the IPv4 CIDR block of the specified subnet.
	CustomVirtualNetworkID string `json:"customVirtualNetworkId"`
}

type AzureSettings struct {
	ResourceGroup string `json:"resourceGroup,omitempty"`
}

// KafkaSpec defines the desired state of Kafka
type KafkaSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// TODO: add comments for fields
	SchemaRegistry []*SchemaRegistry `json:"schema_registry,omitempty"`

	PCIComplianceMode bool `json:"pciComplianceMode"`

	DefaultReplicationFactor int32 `json:"defaultReplicationFactor"`

	DefaultNumberOfPartitions int32 `json:"defaultNumberOfPartitions"`

	RestProxy []*RestProxy `json:"rest_proxy,omitempty"`

	TwoFactorDelete []*TwoFactorDelete `json:"twoFactorDelete,omitempty"`

	AllowDeleteTopics bool `json:"allowDeleteTopics"`

	AutoCreateTopics bool `json:"autoCreateTopics"`

	ClientToClusterEncryption bool `json:"clientToClusterEncryption"`

	KafkaDataCentre []KafkaDataCentre `json:"dataCentres"`

	DedicatedZookeeper []*DedicatedZookeeper `json:"dedicatedZookeeper,omitempty"`

	PrivateNetworkCluster bool `json:"privateNetworkCluster"`

	KafkaVersion string `json:"kafkaVersion"`

	// Name of the cluster.
	Name string `json:"name"`

	SLATier string `json:"slaTier"`
}

// KafkaStatus defines the observed state of Kafka
type KafkaStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ClusterID string `json:"id,omitempty"`

	// ClusterStatus shows cluster current state such as a RUNNING, PROVISIONED, FAILED, etc.
	ClusterStatus string `json:"status,omitempty"`

	Nodes []*Node `json:"nodes"`

	// CurrentClusterOperationStatus indicates if the cluster is currently performing any restructuring operation
	// such as being created or resized. Enum: "NO_OPERATION" "OPERATION_IN_PROGRESS" "OPERATION_FAILED"
	CurrentClusterOperationStatus string `json:"currentClusterOperationStatus"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Kafka is the Schema for the kafkas API
type Kafka struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaSpec   `json:"spec,omitempty"`
	Status KafkaStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KafkaList contains a list of Kafka
type KafkaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kafka `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kafka{}, &KafkaList{})
}
