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

// KafkaSpec defines the desired state of Kafka
type KafkaSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// TODO: add comments for fields
	ClusterSpec               `json:",inline"`
	SchemaRegistry            []*SchemaRegistry     `json:"schema_registry,omitempty"`
	DefaultReplicationFactor  int32                 `json:"defaultReplicationFactor"`
	DefaultNumberOfPartitions int32                 `json:"defaultNumberOfPartitions"`
	RestProxy                 []*RestProxy          `json:"rest_proxy,omitempty"`
	AllowDeleteTopics         bool                  `json:"allowDeleteTopics"`
	AutoCreateTopics          bool                  `json:"autoCreateTopics"`
	ClientToClusterEncryption bool                  `json:"clientToClusterEncryption"`
	KafkaDataCentre           []*DataCentre         `json:"dataCentres"`
	DedicatedZookeeper        []*DedicatedZookeeper `json:"dedicatedZookeeper,omitempty"`
	KafkaVersion              string                `json:"kafkaVersion"`
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