package models

const (
	Redis = "REDIS"
)

type RedisCluster struct {
	Cluster     `json:",inline"`
	Bundles     []*RedisBundle     `json:"bundles"`
	DataCentres []*RedisDataCentre `json:"dataCentres,omitempty"`
}

type RedisBundle struct {
	Bundle  `json:",inline"`
	Options *RedisOptions `json:"options"`
}

type RedisOptions struct {
	ClientEncryption bool  `json:"clientEncryption,omitempty"`
	MasterNodes      int32 `json:"masterNodes"`
	ReplicaNodes     int32 `json:"replicaNodes"`
	PasswordAuth     bool  `json:"passwordAuth,omitempty"`
}

type RedisDataCentre struct {
	DataCentre `json:",inline"`
	Bundles    []*RedisBundle `json:"bundles"`
}
