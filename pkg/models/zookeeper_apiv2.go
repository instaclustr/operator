package models

import modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"

type ZookeeperCluster struct {
	ID                            string                      `json:"ID"`
	Name                          string                      `json:"Name"`
	ZookeeperVersion              string                      `json:"zookeeperVersion"`
	CurrentClusterOperationStatus string                      `json:"currentClusterOperationStatus"`
	Status                        string                      `json:"status"`
	PrivateNetworkCluster         bool                        `json:"privateNetworkCluster"`
	SLATier                       string                      `json:"slaTier"`
	TwoFactorDelete               []*modelsv2.TwoFactorDelete `json:"twoFactorDelete"`
	DataCentres                   []*ZookeeperDataCentre      `json:"dataCentres"`
}

type ZookeeperDataCentre struct {
	modelsv2.DataCentre
	ID                       string           `json:"ID"`
	Status                   string           `json:"status"`
	Nodes                    []*modelsv2.Node `json:"nodes"`
	ClientToServerEncryption bool             `json:"clientToServerEncryption"`
}
