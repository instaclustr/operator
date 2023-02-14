package models

import (
	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
)

type KafkaConnectCluster struct {
	ID                            string                      `json:"ID"`
	Name                          string                      `json:"name"`
	KafkaConnectVersion           string                      `json:"kafkaConnectVersion"`
	Status                        string                      `json:"status"`
	CurrentClusterOperationStatus string                      `json:"currentClusterOperationStatus"`
	PrivateNetworkCluster         bool                        `json:"privateNetworkCluster"`
	SLATier                       string                      `json:"slaTier"`
	TwoFactorDelete               []*modelsv2.TwoFactorDelete `json:"twoFactorDelete"`
	CustomConnectors              []*CustomConnectors         `json:"customConnectors"`
	TargetCluster                 []*TargetCluster            `json:"targetCluster"`
	DataCentres                   []*KafkaConnectDataCentre   `json:"dataCentres"`
}

type ManagedCluster struct {
	TargetKafkaClusterID string `json:"targetKafkaClusterId"`

	// 	Available options are KAFKA_VPC, VPC_PEERED, SEPARATE_VPC
	KafkaConnectVPCType string `json:"kafkaConnectVpcType"`
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

	// Details to connect to a Non-Instaclustr managed cluster. Cannot be provided if targeting an Instaclustr managed cluster.
	ExternalCluster []*ExternalCluster `json:"externalCluster,omitempty"`

	// Details to connect to a Instaclustr managed cluster. Cannot be provided if targeting an external cluster.
	ManagedCluster []*ManagedCluster `json:"managedCluster,omitempty"`
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
	// AWS Identity Access Management role that is used for accessing your specified S3 bucket for Kafka Connect custom connector.
	S3RoleArn string `json:"s3RoleArn"`

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
	modelsv2.DataCentre `json:",inline"`
	ID                  string           `json:"ID"`
	Status              string           `json:"status"`
	ReplicationFactor   int              `json:"replicationFactor"`
	Nodes               []*modelsv2.Node `json:"nodes"`
}
