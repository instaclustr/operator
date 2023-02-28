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
