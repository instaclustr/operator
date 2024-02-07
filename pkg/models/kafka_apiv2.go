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

type KafkaCluster struct {
	GenericClusterFields `json:",inline"`

	KafkaVersion              string `json:"kafkaVersion"`
	AllowDeleteTopics         bool   `json:"allowDeleteTopics"`
	AutoCreateTopics          bool   `json:"autoCreateTopics"`
	BundledUseOnly            bool   `json:"bundledUseOnly"`
	ClientBrokerAuthWithMtls  bool   `json:"clientBrokerAuthWithMtls"`
	ClientToClusterEncryption bool   `json:"clientToClusterEncryption"`
	DefaultNumberOfPartitions int    `json:"defaultNumberOfPartitions"`
	DefaultReplicationFactor  int    `json:"defaultReplicationFactor"`

	DataCentres            []*KafkaDataCentre        `json:"dataCentres"`
	DedicatedZookeeper     []*DedicatedZookeeper     `json:"dedicatedZookeeper"`
	Kraft                  []*Kraft                  `json:"kraft,omitempty"`
	KarapaceRestProxy      []*KarapaceRestProxy      `json:"karapaceRestProxy"`
	KarapaceSchemaRegistry []*KarapaceSchemaRegistry `json:"karapaceSchemaRegistry"`
	RestProxy              []*RestProxy              `json:"restProxy"`
	SchemaRegistry         []*SchemaRegistry         `json:"schemaRegistry"`
	ResizeSettings         []*ResizeSettings         `json:"resizeSettings,omitempty"`
}

type KafkaDataCentre struct {
	GenericDataCentreFields `json:",inline"`

	NumberOfNodes int    `json:"numberOfNodes"`
	NodeSize      string `json:"nodeSize"`

	PrivateLink []*PrivateLink `json:"privateLink,omitempty"`
	Nodes       []*Node        `json:"nodes"`
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

type Kraft struct {
	ControllerNodeCount int `json:"controllerNodeCount"`
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
	ResizeSettings     []*ResizeSettings           `json:"resizeSettings,omitempty"`
}

type DedicatedZookeeperUpdate struct {
	ZookeeperNodeSize string `json:"zookeeperNodeSize"`
}
