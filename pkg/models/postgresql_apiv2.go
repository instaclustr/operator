package models

import (
	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
)

type PGCluster struct {
	Name                  string                      `json:"name"`
	PostgreSQLVersion     string                      `json:"postgresqlVersion"`
	DataCentres           []*PGDataCentre             `json:"dataCentres"`
	SynchronousModeStrict bool                        `json:"synchronousModeStrict"`
	PrivateNetworkCluster bool                        `json:"privateNetworkCluster"`
	SLATier               string                      `json:"slaTier"`
	TwoFactorDelete       []*modelsv2.TwoFactorDelete `json:"twoFactorDelete,omitempty"`
}

type PGBouncer struct {
	PGBouncerVersion string `json:"pgBouncerVersion"`
	PoolMode         string `json:"poolMode"`
}

type PGDataCentre struct {
	modelsv2.DataCentre        `json:",inline"`
	ClientToClusterEncryption  bool                    `json:"clientToClusterEncryption"`
	InterDataCentreReplication []*PGInterDCReplication `json:"interDataCentreReplication,omitempty"`
	IntraDataCentreReplication []*PGIntraDCReplication `json:"intraDataCentreReplication"`
	PGBouncer                  []*PGBouncer            `json:"pgBouncer,omitempty"`
	Status                     string                  `json:"status,omitempty"`
	ID                         string                  `json:"id,omitempty"`
	Nodes                      []*modelsv2.Node        `json:"nodes,omitempty"`
}

type PGInterDCReplication struct {
	IsPrimaryDataCentre bool `json:"isPrimaryDataCentre"`
}

type PGIntraDCReplication struct {
	ReplicationMode string `json:"replicationMode"`
}

type PGStatus struct {
	ID                            string                      `json:"id,omitempty"`
	Name                          string                      `json:"name"`
	PostgreSQLVersion             string                      `json:"postgresqlVersion"`
	PCIComplianceMode             bool                        `json:"PCIComplianceMode,omitempty"`
	DataCentres                   []*PGDataCentre             `json:"dataCentres"`
	CurrentClusterOperationStatus string                      `json:"currentClusterOperationStatus,omitempty"`
	SynchronousModeStrict         bool                        `json:"synchronousModeStrict"`
	PrivateNetworkCluster         bool                        `json:"privateNetworkCluster"`
	TwoFactorDelete               []*modelsv2.TwoFactorDelete `json:"twoFactorDelete,omitempty"`
	SLATier                       string                      `json:"slaTier"`
	Status                        string                      `json:"status,omitempty"`
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
