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
	ClusterStatus
	CassandraVersion      string                 `json:"cassandraVersion"`
	LuceneEnabled         bool                   `json:"luceneEnabled"`
	PasswordAndUserAuth   bool                   `json:"passwordAndUserAuth"`
	Spark                 []*Spark               `json:"spark,omitempty"`
	DataCentres           []*CassandraDataCentre `json:"dataCentres"`
	Name                  string                 `json:"name"`
	SLATier               string                 `json:"slaTier"`
	PrivateNetworkCluster bool                   `json:"privateNetworkCluster"`
	PCIComplianceMode     bool                   `json:"pciComplianceMode"`
	TwoFactorDelete       []*TwoFactorDelete     `json:"twoFactorDelete,omitempty"`
	BundledUseOnly        bool                   `json:"bundledUseOnly,omitempty"`
}

type CassandraDataCentre struct {
	DataCentre                     `json:",inline"`
	ReplicationFactor              int  `json:"replicationFactor"`
	ContinuousBackup               bool `json:"continuousBackup"`
	PrivateIPBroadcastForDiscovery bool `json:"privateIpBroadcastForDiscovery"`
	ClientToClusterEncryption      bool `json:"clientToClusterEncryption"`
}

type Spark struct {
	Version string `json:"version"`
}

type CassandraClusterAPIUpdate struct {
	DataCentres []*CassandraDataCentre `json:"dataCentres"`
}
