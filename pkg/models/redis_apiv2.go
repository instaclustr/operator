package models

import modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"

type RedisCluster struct {
	ID                            string                      `json:"id,omitempty"`
	Name                          string                      `json:"name"`
	RedisVersion                  string                      `json:"redisVersion"`
	ClientToNodeEncryption        bool                        `json:"clientToNodeEncryption"`
	PCIComplianceMode             bool                        `json:"pciComplianceMode"`
	DataCentres                   []*RedisDataCentre          `json:"dataCentres"`
	PrivateNetworkCluster         bool                        `json:"privateNetworkCluster"`
	PasswordAndUserAuth           bool                        `json:"passwordAndUserAuth"`
	TwoFactorDelete               []*modelsv2.TwoFactorDelete `json:"twoFactorDelete,omitempty"`
	SLATier                       string                      `json:"slaTier"`
	CurrentClusterOperationStatus string                      `json:"currentClusterOperationStatus,omitempty"`
	Status                        string                      `json:"status,omitempty"`
}

type RedisDataCentre struct {
	modelsv2.DataCentre
	MasterNodes  int              `json:"masterNodes"`
	ReplicaNodes int              `json:"replicaNodes"`
	ID           string           `json:"id"`
	Status       string           `json:"status"`
	Nodes        []*modelsv2.Node `json:"nodes"`
}

type RedisDataCentreUpdate struct {
	DataCentres []*RedisDataCentre `json:"dataCentres"`
}
