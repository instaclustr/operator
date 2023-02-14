package models

import modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"

type CassandraCluster struct {
	ID                            string                          `json:"id,omitempty"`
	Status                        string                          `json:"status,omitempty"`
	CurrentClusterOperationStatus string                          `json:"currentClusterOperationStatus,omitempty"`
	CassandraVersion              string                          `json:"cassandraVersion"`
	LuceneEnabled                 bool                            `json:"luceneEnabled"`
	PasswordAndUserAuth           bool                            `json:"passwordAndUserAuth"`
	Spark                         []*modelsv2.Spark               `json:"spark,omitempty"`
	DataCentres                   []*modelsv2.CassandraDataCentre `json:"dataCentres"`
	Name                          string                          `json:"name"`
	SLATier                       string                          `json:"slaTier"`
	PrivateNetworkCluster         bool                            `json:"privateNetworkCluster"`
	PCIComplianceMode             bool                            `json:"pciComplianceMode"`
	TwoFactorDeletes              []*modelsv2.TwoFactorDelete     `json:"twoFactorDelete,omitempty"`
}

type CassandraClusterAPIUpdate struct {
	DataCentres []*modelsv2.CassandraDataCentre `json:"dataCentres"`
}
