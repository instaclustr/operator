package models

import (
	"github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
)

type KafkaInstAPICreateRequest struct {
	PCIComplianceMode                 bool                      `json:"pciComplianceMode"`
	SchemaRegistry                    []*SchemaRegistry         `json:"schemaRegistry,omitempty"`
	DefaultReplicationFactor          int32                     `json:"defaultReplicationFactor"`
	DefaultNumberOfPartitions         int32                     `json:"defaultNumberOfPartitions"`
	RestProxy                         []*RestProxy              `json:"restProxy,omitempty"`
	TwoFactorDelete                   []*models.TwoFactorDelete `json:"twoFactorDelete,omitempty"`
	AllowDeleteTopics                 bool                      `json:"allowDeleteTopics"`
	AutoCreateTopics                  bool                      `json:"autoCreateTopics"`
	ClientToClusterEncryption         bool                      `json:"clientToClusterEncryption"`
	KafkaDataCentre                   []*KafkaDataCentre        `json:"dataCentres"`
	DedicatedZookeeper                []*DedicatedZookeeper     `json:"dedicatedZookeeper,omitempty"`
	PrivateNetworkCluster             bool                      `json:"privateNetworkCluster"`
	KafkaVersion                      string                    `json:"kafkaVersion"`
	Name                              string                    `json:"name"`
	SLATier                           string                    `json:"slaTier"`
	ClientBrokerAuthWithMTLS          bool                      `json:"clientBrokerAuthWithMtls,omitempty"`
	ClientAuthBrokerWithoutEncryption bool                      `json:"clientAuthBrokerWithoutEncryption,omitempty"`
	ClientAuthBrokerWithEncryption    bool                      `json:"clientAuthBrokerWithEncryption,omitempty"`
	KarapaceRestProxy                 []*KarapaceRestProxy      `json:"karapaceRestProxy,omitempty"`
	KarapaceSchemaRegistry            []*KarapaceSchemaRegistry `json:"karapaceSchemaRegistry,omitempty"`
	BundledUseOnly                    bool                      `json:"bundledUseOnly,omitempty"`
}

type KafkaDataCentre struct {
	models.DataCentre
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
