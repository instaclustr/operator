package convertors

import (
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
)

type CreateInstaKafkaConnect struct {
	TwoFactorDelete       []*models.TwoFactorDelete    `json:"twoFactorDelete,omitempty"`
	DataCentre            []*KafkaConnectDataCentre    `json:"dataCentres"`
	PrivateNetworkCluster bool                         `json:"privateNetworkCluster"`
	Version               string                       `json:"kafkaConnectVersion"`
	Name                  string                       `json:"name"`
	SLATier               string                       `json:"slaTier"`
	TargetCluster         []*v1alpha1.TargetCluster    `json:"targetCluster"`
	CustomConnectors      []*v1alpha1.CustomConnectors `json:"customConnectors,omitempty"`
}

type KafkaConnectDataCentre struct {
	models.DataCentre `json:",inline"`
	ReplicationFactor int32 `json:"replicationFactor"`
}

func KafkaConnectToInstAPI(crd v1alpha1.KafkaConnectSpec) CreateInstaKafkaConnect {
	return CreateInstaKafkaConnect{
		TwoFactorDelete:       twoFactorDeleteToInstAPI(crd.TwoFactorDelete),
		DataCentre:            kafkaConnectDCsToInstAPI(crd.DataCentres),
		PrivateNetworkCluster: crd.PrivateNetworkCluster,
		Version:               crd.Version,
		Name:                  crd.Name,
		SLATier:               crd.SLATier,
		TargetCluster:         crd.TargetCluster,
		CustomConnectors:      crd.CustomConnectors,
	}
}

func kafkaConnectDCsToInstAPI(crdDCs []*v1alpha1.DataCentre) []*KafkaConnectDataCentre {
	var instaDCs []*KafkaConnectDataCentre

	for _, crdDC := range crdDCs {
		genericDC := &models.DataCentre{
			Name:                crdDC.Name,
			Network:             crdDC.Network,
			NodeSize:            crdDC.NodeSize,
			NumberOfNodes:       crdDC.NodesNumber,
			Tags:                TagsToInstAPI(crdDC.Tags),
			CloudProvider:       crdDC.CloudProvider,
			Region:              crdDC.Region,
			ProviderAccountName: crdDC.ProviderAccountName,
		}
		allocateProviderSettingsToInstAPI(crdDC, genericDC)

		instaDC := &KafkaConnectDataCentre{
			DataCentre:        *genericDC,
			ReplicationFactor: crdDC.RacksNumber,
		}
		instaDCs = append(instaDCs, instaDC)
	}

	return instaDCs
}

type updateKafkaConnectInstAPI struct {
	DataCentre []*KafkaConnectDataCentre `json:"dataCentres"`
}

func KafkaConnectToInstAPIUpdate(k []*v1alpha1.DataCentre) *updateKafkaConnectInstAPI {
	return &updateKafkaConnectInstAPI{
		DataCentre: kafkaConnectDCsToInstAPI(k),
	}
}
