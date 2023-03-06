package models

type ZookeeperCluster struct {
	ID                            string                 `json:"id,omitempty"`
	Name                          string                 `json:"name,omitempty"`
	ZookeeperVersion              string                 `json:"zookeeperVersion,omitempty"`
	CurrentClusterOperationStatus string                 `json:"currentClusterOperationStatus,omitempty"`
	Status                        string                 `json:"status,omitempty"`
	PrivateNetworkCluster         bool                   `json:"privateNetworkCluster"`
	SLATier                       string                 `json:"slaTier,omitempty"`
	TwoFactorDelete               []*TwoFactorDelete     `json:"twoFactorDelete,omitempty"`
	DataCentres                   []*ZookeeperDataCentre `json:"dataCentres,omitempty"`
}

type ZookeeperDataCentre struct {
	DataCentre               `json:",inline"`
	ClientToServerEncryption bool `json:"clientToServerEncryption"`
}
