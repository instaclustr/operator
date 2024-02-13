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

package models

type KafkaConnectCluster struct {
	GenericClusterFields `json:",inline"`

	KafkaConnectVersion   string                    `json:"kafkaConnectVersion"`
	PrivateNetworkCluster bool                      `json:"privateNetworkCluster"`
	CustomConnectors      []*CustomConnectors       `json:"customConnectors,omitempty"`
	TargetCluster         []*TargetCluster          `json:"targetCluster"`
	DataCentres           []*KafkaConnectDataCentre `json:"dataCentres"`
	ResizeSettings        []*ResizeSettings         `json:"resizeSettings,omitempty"`
}

type ManagedCluster struct {
	TargetKafkaClusterID string `json:"targetKafkaClusterId"`
	KafkaConnectVPCType  string `json:"kafkaConnectVpcType"`
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

type TargetCluster struct {
	ExternalCluster []*ExternalCluster `json:"externalCluster,omitempty"`
	ManagedCluster  []*ManagedCluster  `json:"managedCluster,omitempty"`
}

type CustomConnectors struct {
	AzureConnectorSettings []*AzureConnectorSettings `json:"azureConnectorSettings,omitempty"`
	AWSConnectorSettings   []*AWSConnectorSettings   `json:"awsConnectorSettings,omitempty"`
	GCPConnectorSettings   []*GCPConnectorSettings   `json:"gcpConnectorSettings,omitempty"`
}

type AzureConnectorSettings struct {
	StorageContainerName string `json:"storageContainerName"`
	StorageAccountName   string `json:"storageAccountName"`
	StorageAccountKey    string `json:"storageAccountKey"`
}

type AWSConnectorSettings struct {
	S3RoleArn    string `json:"s3RoleArn,omitempty"`
	SecretKey    string `json:"secretKey,omitempty"`
	AccessKey    string `json:"accessKey,omitempty"`
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
	GenericDataCentreFields `json:",inline"`

	NodeSize          string `json:"nodeSize"`
	NumberOfNodes     int    `json:"numberOfNodes"`
	ReplicationFactor int    `json:"replicationFactor"`

	Nodes []*Node `json:"nodes,omitempty"`
}

type KafkaConnectAPIUpdate struct {
	DataCentres    []*KafkaConnectDataCentre `json:"dataCentres,omitempty"`
	ResizeSettings []*ResizeSettings         `json:"resizeSettings,omitempty"`
}
