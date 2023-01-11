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
}

type PGInterDCReplication struct {
	IsPrimaryDataCentre bool `json:"isPrimaryDataCentre"`
}

type PGIntraDCReplication struct {
	ReplicationMode string `json:"replicationMode"`
}
