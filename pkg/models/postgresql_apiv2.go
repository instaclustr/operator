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

type PGCluster struct {
	GenericClusterFields `json:",inline"`

	PostgreSQLVersion     string             `json:"postgresqlVersion"`
	DefaultUserPassword   string             `json:"defaultUserPassword,omitempty"`
	PCIComplianceMode     bool               `json:"pciComplianceMode,omitempty"`
	SynchronousModeStrict bool               `json:"synchronousModeStrict"`
	DataCentres           []*PGDataCentre    `json:"dataCentres"`
	TwoFactorDelete       []*TwoFactorDelete `json:"twoFactorDelete,omitempty"`
	Extensions            []*PGExtension     `json:"extensions,omitempty"`
}

type PGBouncer struct {
	PGBouncerVersion string `json:"pgBouncerVersion"`
	PoolMode         string `json:"poolMode"`
}

type PGDataCentre struct {
	GenericDataCentreFields `json:",inline"`

	NumberOfNodes             int    `json:"numberOfNodes"`
	NodeSize                  string `json:"nodeSize"`
	ClientToClusterEncryption bool   `json:"clientToClusterEncryption"`

	InterDataCentreReplication []*PGInterDCReplication `json:"interDataCentreReplication,omitempty"`
	IntraDataCentreReplication []*PGIntraDCReplication `json:"intraDataCentreReplication"`
	PGBouncer                  []*PGBouncer            `json:"pgBouncer,omitempty"`
	Nodes                      []*Node                 `json:"nodes,omitempty"`
}

type PGInterDCReplication struct {
	IsPrimaryDataCentre bool `json:"isPrimaryDataCentre"`
}

type PGIntraDCReplication struct {
	ReplicationMode string `json:"replicationMode"`
}

type PGConfigs struct {
	ClusterID               string                     `json:"clusterId,omitempty"`
	ConfigurationProperties []*ConfigurationProperties `json:"configurationProperties"`
}

type ConfigurationProperties struct {
	Name      string `json:"name"`
	ClusterID string `json:"clusterId"`
	ID        string `json:"id,omitempty"`
	Value     string `json:"value"`
}

type PGClusterUpdate struct {
	DataCentres    []*PGDataCentre   `json:"dataCentres"`
	ResizeSettings []*ResizeSettings `json:"resizeSettings,omitempty"`
}

type PGExtension struct {
	Name    string `json:"extensionName"`
	Enabled bool   `json:"extensionEnabled"`
}
