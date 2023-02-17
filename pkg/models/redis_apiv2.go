package models

type RedisCluster struct {
	ClusterStatus          `json:",inline"`
	Name                   string             `json:"name,omitempty"`
	RedisVersion           string             `json:"redisVersion,omitempty"`
	ClientToNodeEncryption bool               `json:"clientToNodeEncryption,omitempty"`
	PCIComplianceMode      bool               `json:"pciComplianceMode,omitempty"`
	DataCentres            []*RedisDataCentre `json:"dataCentres,omitempty"`
	PrivateNetworkCluster  bool               `json:"privateNetworkCluster,omitempty"`
	PasswordAndUserAuth    bool               `json:"passwordAndUserAuth,omitempty"`
	TwoFactorDelete        []*TwoFactorDelete `json:"twoFactorDelete,omitempty"`
	SLATier                string             `json:"slaTier,omitempty"`
}

type RedisDataCentre struct {
	DataCentre   `json:",inline"`
	MasterNodes  int `json:"masterNodes"`
	ReplicaNodes int `json:"replicaNodes"`
}

type RedisDataCentreUpdate struct {
	DataCentres []*RedisDataCentre `json:"dataCentres"`
}
