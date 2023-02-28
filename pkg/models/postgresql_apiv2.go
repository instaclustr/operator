package models

type PGCluster struct {
	ID                            string             `json:"id,omitempty"`
	Name                          string             `json:"name"`
	PostgreSQLVersion             string             `json:"postgresqlVersion"`
	DataCentres                   []*PGDataCentre    `json:"dataCentres"`
	SynchronousModeStrict         bool               `json:"synchronousModeStrict"`
	PrivateNetworkCluster         bool               `json:"privateNetworkCluster"`
	SLATier                       string             `json:"slaTier"`
	TwoFactorDelete               []*TwoFactorDelete `json:"twoFactorDelete,omitempty"`
	PCIComplianceMode             bool               `json:"pciComplianceMode,omitempty"`
	CurrentClusterOperationStatus string             `json:"currentClusterOperationStatus,omitempty"`
	Status                        string             `json:"status,omitempty"`
}

type PGBouncer struct {
	PGBouncerVersion string `json:"pgBouncerVersion"`
	PoolMode         string `json:"poolMode"`
}

type PGDataCentre struct {
	DataCentre                 `json:",inline"`
	ClientToClusterEncryption  bool                    `json:"clientToClusterEncryption"`
	InterDataCentreReplication []*PGInterDCReplication `json:"interDataCentreReplication,omitempty"`
	IntraDataCentreReplication []*PGIntraDCReplication `json:"intraDataCentreReplication"`
	PGBouncer                  []*PGBouncer            `json:"pgBouncer,omitempty"`
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
