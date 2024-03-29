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

type RedisCluster struct {
	GenericClusterFields `json:",inline"`

	RedisVersion           string `json:"redisVersion"`
	ClientToNodeEncryption bool   `json:"clientToNodeEncryption"`
	PasswordAndUserAuth    bool   `json:"passwordAndUserAuth"`
	PCIComplianceMode      bool   `json:"pciComplianceMode"`

	DataCentres []*RedisDataCentre `json:"dataCentres,omitempty"`
}

type RedisDataCentre struct {
	GenericDataCentreFields `json:",inline"`

	NodeSize          string `json:"nodeSize"`
	MasterNodes       int    `json:"masterNodes"`
	ReplicaNodes      int    `json:"replicaNodes"`
	ReplicationFactor int    `json:"replicationFactor"`

	Nodes       []*Node        `json:"nodes,omitempty"`
	PrivateLink []*PrivateLink `json:"privateLink,omitempty"`
}

type RedisDataCentreUpdate struct {
	DataCentres    []*RedisDataCentre `json:"dataCentres"`
	ResizeSettings []*ResizeSettings  `json:"resizeSettings,omitempty"`
}

type RedisUser struct {
	ID                 string `json:"ID,omitempty"`
	ClusterID          string `json:"clusterId"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	InitialPermissions string `json:"initialPermissions"`
}

type RedisUserUpdate struct {
	ID       string
	Password string `json:"password"`
}
