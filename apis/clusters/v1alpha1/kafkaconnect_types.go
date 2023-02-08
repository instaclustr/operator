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

type TargetCluster struct {

	// Details to connect to a Non-Instaclustr managed cluster. Cannot be provided if targeting an Instaclustr managed cluster.
	ExternalCluster []*ExternalCluster `json:"externalCluster,omitempty"`

	// Details to connect to a Instaclustr managed cluster. Cannot be provided if targeting an external cluster.
	ManagedCluster []*ManagedCluster `json:"managedCluster,omitempty"`
}

type ExternalCluster struct {
	SecurityProtocol      string `json:"securityProtocol,omitempty"`
	SSLTruststorePassword string `json:"sslTruststorePassword,omitempty"`
	BootstrapServers      string `json:"bootstrapServers,omitempty"`
	SASLJAASConfig        string `json:"saslJaasConfig,omitempty"`
	SASLMechanism         string `json:"saslMechanism,omitempty"`
	SSLProtocol           string `json:"sslProtocol,omitempty"`
	SSLEnabledProtocols   string `json:"sslEnabledProtocols,omitempty"`
	Truststore            string `json:"truststore,omitempty"`
}

type ManagedCluster struct {
	TargetKafkaClusterID string `json:"targetKafkaClusterId"`

	// 	Available options are KAFKA_VPC, VPC_PEERED, SEPARATE_VPC
	KafkaConnectVPCType string `json:"kafkaConnectVpcType"`
}

type CustomConnectors struct {
	// Settings to access custom connectors located in an azure storage container.
	AzureConnectorSettings []*AzureConnectorSettings `json:"azureConnectorSettings,omitempty"`

	// Settings to access custom connectors located in a S3 bucket.
	AWSConnectorSettings []*AWSConnectorSettings `json:"awsConnectorSettings,omitempty"`

	// Settings to access custom connectors located in a gcp storage container.
	GCPConnectorSettings []*GCPConnectorSettings `json:"gcpConnectorSettings,omitempty"`
}

type AzureConnectorSettings struct {
	// Azure storage container name for Kafka Connect custom connector.
	StorageContainerName string `json:"storageContainerName"`

	// Azure storage account name to access your Azure bucket for Kafka Connect custom connector.
	StorageAccountName string `json:"storageAccountName"`

	// Azure storage account key to access your Azure bucket for Kafka Connect custom connector.
	StorageAccountKey string `json:"storageAccountKey"`
}

type AWSConnectorSettings struct {
	// AWS Secret Key associated with the Access Key id that can access your specified S3 bucket for Kafka Connect custom connector.
	SecretKey string `json:"secretKey"`

	// AWS Access Key id that can access your specified S3 bucket for Kafka Connect custom connector.
	AccessKey string `json:"accessKey"`

	// S3 bucket name for Kafka Connect custom connector.
	S3BucketName string `json:"s3BucketName"`
}

type GCPConnectorSettings struct {
	PrivateKey        string `json:"privateKey"`
	ClientID          string `json:"clientId"`
	ClientEmail       string `json:"clientEmail"`
	ProjectID         string `json:"projectId"`
	StorageBucketName string `json:"storageBucketName"`
	PrivateKeyID      string `json:"privateKeyId"`
}

type KafkaConnectDataCentre struct {
	DataCentre        `json:",inline"`
	ReplicationFactor int32 `json:"replicationFactor"`
}

// KafkaConnectSpec defines the desired state of KafkaConnect
type KafkaConnectSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Cluster       `json:",inline"`
	DataCentres   []*KafkaConnectDataCentre `json:"dataCentres"`
	TargetCluster []*TargetCluster          `json:"targetCluster"`

	// CustomConnectors defines the location for custom connector storage and access info.
	CustomConnectors []*CustomConnectors `json:"customConnectors,omitempty"`
}

// KafkaConnectStatus defines the observed state of KafkaConnect
type KafkaConnectStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ClusterStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KafkaConnect is the Schema for the kafkaconnects API
type KafkaConnect struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KafkaConnectSpec   `json:"spec,omitempty"`
	Status KafkaConnectStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KafkaConnectList contains a list of KafkaConnect
type KafkaConnectList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KafkaConnect `json:"items"`
}

func (k *KafkaConnect) GetJobID(jobName string) string {
	return client.ObjectKeyFromObject(k).String() + "/" + jobName
}

func (k *KafkaConnect) NewPatch() client.Patch {
	old := k.DeepCopy()
	old.Annotations[models.ResourceStateAnnotation] = ""
	return client.MergeFrom(old)
}

func init() {
	SchemeBuilder.Register(&KafkaConnect{}, &KafkaConnectList{})
}
