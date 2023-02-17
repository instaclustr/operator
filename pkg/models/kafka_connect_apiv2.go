package models

type KafkaConnectCluster struct {
	ClusterStatus         `json:",inline"`
	Name                  string                    `json:"name,omitempty"`
	KafkaConnectVersion   string                    `json:"kafkaConnectVersion,omitempty"`
	PrivateNetworkCluster bool                      `json:"privateNetworkCluster,omitempty"`
	SLATier               string                    `json:"slaTier,omitempty"`
	TwoFactorDelete       []*TwoFactorDelete        `json:"twoFactorDelete,omitempty"`
	CustomConnectors      []*CustomConnectors       `json:"customConnectors,omitempty"`
	TargetCluster         []*TargetCluster          `json:"targetCluster,omitempty"`
	DataCentres           []*KafkaConnectDataCentre `json:"dataCentres,omitempty"`
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
	S3RoleArn    string `json:"s3RoleArn"`
	SecretKey    string `json:"secretKey"`
	AccessKey    string `json:"accessKey"`
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
	ReplicationFactor int `json:"replicationFactor"`
}

type KafkaConnectAPIUpdate struct {
	DataCentres []*KafkaConnectDataCentre `json:"dataCentres,omitempty"`
}
