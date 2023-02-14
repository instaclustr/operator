package models

import (
	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
)

type KafkaCluster struct {
	ID                                string                      `json:"ID"`
	Status                            string                      `json:"status"`
	CurrentClusterOperationStatus     string                      `json:"currentClusterOperationStatus"`
	Name                              string                      `json:"name"`
	KafkaVersion                      string                      `json:"kafkaVersion"`
	PrivateNetworkCluster             bool                        `json:"privateNetworkCluster"`
	SLATier                           string                      `json:"slaTier"`
	TwoFactorDelete                   []*modelsv2.TwoFactorDelete `json:"twoFactorDelete"`
	AllowDeleteTopics                 bool                        `json:"allowDeleteTopics"`
	AutoCreateTopics                  bool                        `json:"autoCreateTopics"`
	BundledUseOnly                    bool                        `json:"bundledUseOnly"`
	ClientAuthBrokerWithEncryption    bool                        `json:"clientAuthBrokerWithEncryption"`
	ClientAuthBrokerWithoutEncryption bool                        `json:"clientAuthBrokerWithoutEncryption"`
	ClientBrokerAuthWithMtls          bool                        `json:"clientBrokerAuthWithMtls"`
	ClientToClusterEncryption         bool                        `json:"clientToClusterEncryption"`
	DataCentres                       []*KafkaDataCentre          `json:"dataCentres"`
	DedicatedZookeeper                []*DedicatedZookeeper       `json:"dedicatedZookeeper"`
	DefaultNumberOfPartitions         int                         `json:"defaultNumberOfPartitions"`
	DefaultReplicationFactor          int                         `json:"defaultReplicationFactor"`
	KarapaceRestProxy                 []*KarapaceRestProxy        `json:"karapaceRestProxy"`
	KarapaceSchemaRegistry            []*KarapaceSchemaRegistry   `json:"karapaceSchemaRegistry"`
	PCIComplianceMode                 bool                        `json:"PCIComplianceMode"`
	RestProxy                         []*RestProxy                `json:"restProxy"`
	SchemaRegistry                    []*SchemaRegistry           `json:"schemaRegistry"`
}

type KafkaInstAPICreateRequest struct {
	PCIComplianceMode                 bool                        `json:"pciComplianceMode"`
	SchemaRegistry                    []*SchemaRegistry           `json:"schemaRegistry,omitempty"`
	DefaultReplicationFactor          int32                       `json:"defaultReplicationFactor"`
	DefaultNumberOfPartitions         int32                       `json:"defaultNumberOfPartitions"`
	RestProxy                         []*RestProxy                `json:"restProxy,omitempty"`
	TwoFactorDelete                   []*modelsv2.TwoFactorDelete `json:"twoFactorDelete,omitempty"`
	AllowDeleteTopics                 bool                        `json:"allowDeleteTopics"`
	AutoCreateTopics                  bool                        `json:"autoCreateTopics"`
	ClientToClusterEncryption         bool                        `json:"clientToClusterEncryption"`
	KafkaDataCentre                   []*KafkaDataCentre          `json:"dataCentres"`
	DedicatedZookeeper                []*DedicatedZookeeper       `json:"dedicatedZookeeper,omitempty"`
	PrivateNetworkCluster             bool                        `json:"privateNetworkCluster"`
	KafkaVersion                      string                      `json:"kafkaVersion"`
	Name                              string                      `json:"name"`
	SLATier                           string                      `json:"slaTier"`
	ClientBrokerAuthWithMTLS          bool                        `json:"clientBrokerAuthWithMtls,omitempty"`
	ClientAuthBrokerWithoutEncryption bool                        `json:"clientAuthBrokerWithoutEncryption,omitempty"`
	ClientAuthBrokerWithEncryption    bool                        `json:"clientAuthBrokerWithEncryption,omitempty"`
	KarapaceRestProxy                 []*KarapaceRestProxy        `json:"karapaceRestProxy,omitempty"`
	KarapaceSchemaRegistry            []*KarapaceSchemaRegistry   `json:"karapaceSchemaRegistry,omitempty"`
	BundledUseOnly                    bool                        `json:"bundledUseOnly,omitempty"`
}

type KafkaDataCentre struct {
	ID                  string           `json:"ID,omitempty,"`
	Status              string           `json:"status,omitempty"`
	Nodes               []*modelsv2.Node `json:"nodes,omitempty"`
	modelsv2.DataCentre `json:",inline"`
	PrivateLink         []*KafkaPrivateLink `json:"privateLink,omitempty"`
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
