package models

import modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"

type RedisCluster struct {
	Name                   string                      `json:"name"`
	RedisVersion           string                      `json:"redisVersion"`
	DataCentres            []*RedisDataCentre          `json:"dataCentres"`
	SLATier                string                      `json:"slaTier"`
	ClientToNodeEncryption bool                        `json:"clientToNodeEncryption"`
	PCIComplianceMode      bool                        `json:"pciComplianceMode"`
	PrivateNetworkCluster  bool                        `json:"privateNetworkCluster"`
	PasswordAndUserAuth    bool                        `json:"passwordAndUserAuth"`
	TwoFactorDelete        []*modelsv2.TwoFactorDelete `json:"twoFactorDelete,omitempty"`
}

type RedisDataCentre struct {
	modelsv2.DataCentre `json:",inline"`
	RedisVersion        string `json:"redisVersion"`
	MasterNodes         int    `json:"masterNodes"`
	ReplicaNodes        int    `json:"replicaNodes"`
}
