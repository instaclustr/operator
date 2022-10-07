package models

type CreateKafka struct {
	PCIComplianceMode         bool                  `json:"pciComplianceMode"`
	SchemaRegistry            []*SchemaRegistry     `json:"schemaRegistry,omitempty"`
	DefaultReplicationFactor  int32                 `json:"defaultReplicationFactor"`
	DefaultNumberOfPartitions int32                 `json:"defaultNumberOfPartitions"`
	RestProxy                 []*RestProxy          `json:"restProxy,omitempty"`
	TwoFactorDelete           []*TwoFactorDelete    `json:"twoFactorDelete,omitempty"`
	AllowDeleteTopics         bool                  `json:"allowDeleteTopics"`
	AutoCreateTopics          bool                  `json:"autoCreateTopics"`
	ClientToClusterEncryption bool                  `json:"clientToClusterEncryption"`
	KafkaDataCentre           []*DataCentre         `json:"dataCentres"`
	DedicatedZookeeper        []*DedicatedZookeeper `json:"dedicatedZookeeper,omitempty"`
	PrivateNetworkCluster     bool                  `json:"privateNetworkCluster"`
	KafkaVersion              string                `json:"kafkaVersion"`
	Name                      string                `json:"name"`
	SLATier                   string                `json:"slaTier"`
}

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
