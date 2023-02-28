package models

type KafkaCluster struct {
	ClusterStatus                     `json:",inline"`
	Name                              string                    `json:"name"`
	KafkaVersion                      string                    `json:"kafkaVersion"`
	PrivateNetworkCluster             bool                      `json:"privateNetworkCluster"`
	SLATier                           string                    `json:"slaTier"`
	TwoFactorDelete                   []*TwoFactorDelete        `json:"twoFactorDelete"`
	AllowDeleteTopics                 bool                      `json:"allowDeleteTopics"`
	AutoCreateTopics                  bool                      `json:"autoCreateTopics"`
	BundledUseOnly                    bool                      `json:"bundledUseOnly"`
	ClientAuthBrokerWithEncryption    bool                      `json:"clientAuthBrokerWithEncryption"`
	ClientAuthBrokerWithoutEncryption bool                      `json:"clientAuthBrokerWithoutEncryption"`
	ClientBrokerAuthWithMtls          bool                      `json:"clientBrokerAuthWithMtls"`
	ClientToClusterEncryption         bool                      `json:"clientToClusterEncryption"`
	DataCentres                       []*KafkaDataCentre        `json:"dataCentres"`
	DedicatedZookeeper                []*DedicatedZookeeper     `json:"dedicatedZookeeper"`
	DefaultNumberOfPartitions         int                       `json:"defaultNumberOfPartitions"`
	DefaultReplicationFactor          int                       `json:"defaultReplicationFactor"`
	KarapaceRestProxy                 []*KarapaceRestProxy      `json:"karapaceRestProxy"`
	KarapaceSchemaRegistry            []*KarapaceSchemaRegistry `json:"karapaceSchemaRegistry"`
	PCIComplianceMode                 bool                      `json:"pciComplianceMode"`
	RestProxy                         []*RestProxy              `json:"restProxy"`
	SchemaRegistry                    []*SchemaRegistry         `json:"schemaRegistry"`
}

type KafkaDataCentre struct {
	DataCentre  `json:",inline"`
	PrivateLink []*KafkaPrivateLink `json:"privateLink,omitempty"`
}

type KafkaPrivateLink struct {
	AdvertisedHostname string `json:"advertisedHostname"`
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

type KarapaceRestProxy struct {
	IntegrateRestProxyWithSchemaRegistry bool   `json:"integrateRestProxyWithSchemaRegistry"`
	Version                              string `json:"version"`
}

type KarapaceSchemaRegistry struct {
	Version string `json:"version"`
}

type DedicatedZookeeper struct {
	ZookeeperNodeSize  string `json:"zookeeperNodeSize"`
	ZookeeperNodeCount int32  `json:"zookeeperNodeCount"`
}

type KafkaInstAPIUpdateRequest struct {
	DataCentre         []*KafkaDataCentre          `json:"dataCentres"`
	DedicatedZookeeper []*DedicatedZookeeperUpdate `json:"dedicatedZookeeper,omitempty"`
}

type DedicatedZookeeperUpdate struct {
	ZookeeperNodeSize string `json:"zookeeperNodeSize"`
}
