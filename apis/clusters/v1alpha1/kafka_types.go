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

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/models"
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
	// Size of the nodes provisioned as dedicated Zookeeper nodes.
	NodeSize string `json:"nodeSize"`

	// Number of dedicated Zookeeper node count, it must be 3 or 5.
	NodesNumber int32 `json:"nodesNumber"`
}

type KarapaceRestProxy struct {
	IntegrateRestProxyWithSchemaRegistry bool   `json:"integrateRestProxyWithSchemaRegistry"`
	Version                              string `json:"version"`
}

type KarapaceSchemaRegistry struct {
	Version string `json:"version"`
}

// KafkaSpec defines the desired state of Kafka
type KafkaSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Cluster        `json:",inline"`
	SchemaRegistry []*SchemaRegistry `json:"schemaRegistry,omitempty"`

	// ReplicationFactorNumber to use for new topic.
	// Also represents the number of racks to use when allocating nodes.
	ReplicationFactorNumber int32 `json:"replicationFactorNumber"`

	// PartitionsNumber number of partitions to use when created new topics.
	PartitionsNumber          int32              `json:"partitionsNumber"`
	RestProxy                 []*RestProxy       `json:"restProxy,omitempty"`
	AllowDeleteTopics         bool               `json:"allowDeleteTopics"`
	AutoCreateTopics          bool               `json:"autoCreateTopics"`
	ClientToClusterEncryption bool               `json:"clientToClusterEncryption"`
	DataCentres               []*KafkaDataCentre `json:"dataCentres"`

	// Provision additional dedicated nodes for Apache Zookeeper to run on.
	// Zookeeper nodes will be co-located with Kafka if this is not provided
	DedicatedZookeeper                []*DedicatedZookeeper     `json:"dedicatedZookeeper,omitempty"`
	ClientBrokerAuthWithMTLS          bool                      `json:"clientBrokerAuthWithMtls,omitempty"`
	ClientAuthBrokerWithoutEncryption bool                      `json:"clientAuthBrokerWithoutEncryption,omitempty"`
	ClientAuthBrokerWithEncryption    bool                      `json:"clientAuthBrokerWithEncryption,omitempty"`
	KarapaceRestProxy                 []*KarapaceRestProxy      `json:"karapaceRestProxy,omitempty"`
	KarapaceSchemaRegistry            []*KarapaceSchemaRegistry `json:"karapaceSchemaRegistry,omitempty"`
}

type KafkaDataCentre struct {
	Name                  string                   `json:"name,omitempty"`
	Region                string                   `json:"region"`
	CloudProvider         string                   `json:"cloudProvider"`
	ProviderAccountName   string                   `json:"accountName,omitempty"`
	CloudProviderSettings []*CloudProviderSettings `json:"cloudProviderSettings,omitempty"`
	Network               string                   `json:"network"`
	NodeSize              string                   `json:"nodeSize"`
	NodesNumber           int32                    `json:"nodesNumber"`
	Tags                  map[string]string        `json:"tags,omitempty"`
}

// KafkaStatus defines the observed state of Kafka
type KafkaStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ClusterStatus `json:",inline"`

	// +optional
	// CurrentClusterOperation indicates if the cluster is currently performing any restructuring operation
	// such as being created or resized. Enum: "NO_OPERATION" "OPERATION_IN_PROGRESS" "OPERATION_FAILED"
	CurrentClusterOperation string `json:"currentClusterOperationStatus"`
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

func (k *Kafka) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(k).String() + "/" + jobName
}

func (k *Kafka) NewPatch() client.Patch {
	old := k.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func init() {
	SchemeBuilder.Register(&Kafka{}, &KafkaList{})
}
