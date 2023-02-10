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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/instaclustr/operator/pkg/models"
)

type SourceCluster struct {
	// Alias to use for the source kafka cluster. This will be used to rename topics if renameMirroredTopics is turned on
	Alias string `json:"alias"`

	// Details to connect to a Non-Instaclustr managed cluster. Cannot be provided if targeting an Instaclustr managed cluster
	ExternalCluster []*ExternalCluster `json:"externalCluster,omitempty"`

	// Details to connect to a Instaclustr managed cluster. Cannot be provided if targeting a non-Instaclustr managed cluster
	ManagedCluster []*ManagedCluster `json:"managedCluster,omitempty"`
}

type ExternalCluster struct {
	// Kafka connection properties string used to connect to external kafka cluster
	SourceConnectionProperties string `json:"sourceConnectionProperties"`
}

type ManagedCluster struct {
	// Whether or not to connect to source cluster's private IP addresses
	UsePrivateIPs bool `json:"usePrivateIps"`

	// Source kafka cluster id
	SourceKafkaClusterID string `json:"sourceKafkaClusterId"`
}

// MirrorSpec defines the desired state of Mirror
type MirrorSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The latency in milliseconds above which this mirror will be considered out of sync. It can not be less than 1000ms
	// The suggested initial latency is 30000ms for connectors to be created.
	TargetLatency int32 `json:"targetLatency"`

	// ID of the kafka connect cluster
	KafkaConnectClusterID string `json:"kafkaConnectClusterId"`

	// Maximum number of tasks for Kafka Connect to use. Should be greater than 0
	MaxTasks int32 `json:"maxTasks"`

	// Details to connect to the source kafka cluster
	SourceCluster []*SourceCluster `json:"sourceCluster"`

	// Whether to rename topics as they are mirrored, by prefixing the sourceCluster.alias to the topic name
	RenameMirroredTopics bool `json:"renameMirroredTopics"`

	// Regex to select which topics to mirror
	TopicsRegex string `json:"topicsRegex"`
}

type Connector struct {
	// Name of the connector. Could be one of [Mirror Connector, Checkpoint Connector].
	Name string `json:"name"`

	// Configuration of the connector.
	Config string `json:"config"`

	// Status of the connector.
	Status string `json:"status"`
}

type MirroredTopic struct {
	// Average latency in milliseconds for messages to travel from source to destination topics.
	AverageLatency resource.Quantity `json:"averageLatency"`

	// Average record rate for messages to travel from source to destination topics, it is 0 if there are no messages travelling in between.
	AverageRate resource.Quantity `json:"averageRate"`

	// Name of the mirrored topic.
	Name string `json:"name"`
}

// MirrorStatus defines the observed state of Mirror
type MirrorStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Detailed list of Connectors for the mirror.
	Connectors []*Connector `json:"connectors"`

	// Name of the mirror connector. The value of this property has the form: [source-cluster].[target-cluster].[random-string]
	ConnectorName string `json:"connectorName"`

	// ID of the mirror
	ID string `json:"id"`

	// Detailed list of Mirrored topics.
	MirroredTopics []*MirroredTopic `json:"mirroredTopics"`

	// The overall status of this mirror.
	Status string `json:"status"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Mirror is the Schema for the mirrors API
type Mirror struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MirrorSpec   `json:"spec,omitempty"`
	Status MirrorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MirrorList contains a list of Mirror
type MirrorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Mirror `json:"items"`
}

func (m *Mirror) NewPatch() client.Patch {
	old := m.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func (m *Mirror) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(m).String() + "/" + jobName
}

func init() {
	SchemeBuilder.Register(&Mirror{}, &MirrorList{})
}
