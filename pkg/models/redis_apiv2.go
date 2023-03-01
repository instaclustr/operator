package models

type RedisCluster struct {
	ClusterStatus          `json:",inline"`
	Name                   string             `json:"name"`
	RedisVersion           string             `json:"redisVersion"`
	ClientToNodeEncryption bool               `json:"clientToNodeEncryption"`
	PCIComplianceMode      bool               `json:"pciComplianceMode"`
	DataCentres            []*RedisDataCentre `json:"dataCentres,omitempty"`
	PrivateNetworkCluster  bool               `json:"privateNetworkCluster"`
	PasswordAndUserAuth    bool               `json:"passwordAndUserAuth"`
	TwoFactorDelete        []*TwoFactorDelete `json:"twoFactorDelete,omitempty"`
	SLATier                string             `json:"slaTier"`
}

type RedisDataCentre struct {
	DataCentre   `json:",inline"`
	MasterNodes  int `json:"masterNodes"`
	ReplicaNodes int `json:"replicaNodes"`
}

type RedisDataCentreUpdate struct {
	DataCentres []*RedisDataCentre `json:"dataCentres"`
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
