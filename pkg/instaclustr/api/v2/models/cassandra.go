package models

type CassandraCluster struct {
	ClusterSpec         `json:",inline"`
	DataCentres         []*CassandraDataCentre `json:"dataCentres"`
	CassandraVersion    string                 `json:"cassandraVersion"`
	LuceneEnabled       bool                   `json:"luceneEnabled"`
	PasswordAndUserAuth bool                   `json:"passwordAndUserAuth"`
}

type CassandraDataCentre struct {
	DataCentre                     `json:",inline"`
	ReplicationFactor              int32 `json:"replicationFactor"`
	ContinuousBackup               bool  `json:"continuousBackup"`
	PrivateIPBroadcastForDiscovery bool  `json:"privateIpBroadcastForDiscovery"`
	ClientToClusterEncryption      bool  `json:"clientToClusterEncryption"`
}

// CassandraStatus defines the observed state of Cassandra
type CassandraStatus struct {
	ClusterStatus   `json:",inline"`
	OperationStatus string `json:"operationStatus,omitempty"`
}
