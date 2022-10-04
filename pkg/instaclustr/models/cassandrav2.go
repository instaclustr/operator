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
	ContinuousBackup               bool `json:"continuousBackup"`
	ReplicationFactor              int  `json:"replicationFactor"`
	PrivateIPBroadcastForDiscovery bool `json:"privateIpBroadcastForDiscovery"`
	ClientToClusterEncryption      bool `json:"clientToClusterEncryption"`
}

type CassandraDCStatus struct {
	ID             string  `json:"id,omitempty"`
	Status         string  `json:"status,omitempty"`
	CassandraNodes []*Node `json:"nodes,omitempty"`
}

type CassandraStatus struct {
	ClusterID         string               `json:"id,omitempty"`
	ClusterStatus     string               `json:"status,omitempty"`
	OperationStatus   string               `json:"operationStatus,omitempty"`
	CassandraDCStatus []*CassandraDCStatus `json:"dataCentres,omitempty"`
}
