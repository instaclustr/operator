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

type CassandraCluster struct {
	GenericClusterFields `json:",inline"`

	CassandraVersion    string `json:"cassandraVersion"`
	LuceneEnabled       bool   `json:"luceneEnabled"`
	PasswordAndUserAuth bool   `json:"passwordAndUserAuth"`
	BundledUseOnly      bool   `json:"bundledUseOnly,omitempty"`
	PCIComplianceMode   bool   `json:"pciComplianceMode"`

	DataCentres    []*CassandraDataCentre `json:"dataCentres"`
	ResizeSettings []*ResizeSettings      `json:"resizeSettings"`
}

type CassandraDataCentre struct {
	GenericDataCentreFields `json:",inline"`

	ContinuousBackup               bool   `json:"continuousBackup"`
	PrivateLink                    bool   `json:"privateLink"`
	PrivateIPBroadcastForDiscovery bool   `json:"privateIpBroadcastForDiscovery"`
	ClientToClusterEncryption      bool   `json:"clientToClusterEncryption"`
	ReplicationFactor              int    `json:"replicationFactor"`
	NumberOfNodes                  int    `json:"numberOfNodes"`
	NodeSize                       string `json:"nodeSize"`

	Nodes         []*Node          `json:"nodes,omitempty"`
	Debezium      []*Debezium      `json:"debezium,omitempty"`
	ShotoverProxy []*ShotoverProxy `json:"shotoverProxy,omitempty"`
}

type Debezium struct {
	KafkaVPCType      string `json:"kafkaVpcType"`
	KafkaTopicPrefix  string `json:"kafkaTopicPrefix"`
	KafkaDataCentreID string `json:"kafkaCdcId"`
	Version           string `json:"version"`
}

type ShotoverProxy struct {
	NodeSize string `json:"nodeSize"`
}

type CassandraClusterAPIUpdate struct {
	DataCentres    []*CassandraDataCentre `json:"dataCentres"`
	ResizeSettings []*ResizeSettings      `json:"resizeSettings,omitempty"`
}
