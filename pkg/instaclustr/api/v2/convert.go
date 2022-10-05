package v2

import (
	"encoding/json"

	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
)

func CassandraToInstAPI(cassandraSpec *v1alpha1.CassandraSpec) *models.CassandraCluster {
	cassandraInstTwoFactorDelete := cassandraTwoFactorDeleteToInstAPI(cassandraSpec.TwoFactorDelete)

	cassandra := &models.CassandraCluster{
		ClusterSpec: models.ClusterSpec{
			Name:                  cassandraSpec.Name,
			SLATier:               cassandraSpec.SLATier,
			PrivateNetworkCluster: cassandraSpec.PrivateNetworkCluster,
			PCIComplianceMode:     cassandraSpec.PCICompliance,
			TwoFactorDeletes:      cassandraInstTwoFactorDelete,
		},
		CassandraVersion:    cassandraSpec.Version,
		LuceneEnabled:       cassandraSpec.LuceneEnabled,
		PasswordAndUserAuth: cassandraSpec.PasswordAndUserAuth,
	}

	var cassandraInstDCs []*models.CassandraDataCentre
	for _, dataCentre := range cassandraSpec.DataCentres {

		cassandraInstDC := &models.CassandraDataCentre{
			DataCentre: models.DataCentre{
				Name:                dataCentre.Name,
				Network:             dataCentre.Network,
				NodeSize:            dataCentre.NodeSize,
				NumberOfNodes:       dataCentre.NodesNumber,
				Region:              dataCentre.Region,
				Tags:                tagsToInstAPI(dataCentre.Tags),
				CloudProvider:       dataCentre.CloudProvider,
				ProviderAccountName: dataCentre.AccountName,
			},
			ContinuousBackup:               dataCentre.ContinuousBackup,
			PrivateIPBroadcastForDiscovery: dataCentre.PrivateIPBroadcastForDiscovery,
			ClientToClusterEncryption:      dataCentre.ClientToClusterEncryption,
			ReplicationFactor:              dataCentre.RacksNumber,
		}

		providerSettingsToInstaCloudProviders(cassandraInstDCs, dataCentre)
		cassandraInstDCs = append(cassandraInstDCs, cassandraInstDC)

	}
	cassandra.DataCentres = cassandraInstDCs

	return cassandra
}

func cassandraTwoFactorDeleteToInstAPI(twoFactorDelete []*v1alpha1.TwoFactorDelete) []*models.TwoFactorDelete {
	if len(twoFactorDelete) < 1 {
		return nil
	}

	var twoFactor []*models.TwoFactorDelete
	for i := range twoFactorDelete {
		twoFactor[i].ConfirmationEmail = twoFactorDelete[i].Email
		twoFactor[i].ConfirmationPhoneNumber = twoFactorDelete[i].Phone
	}

	return twoFactor
}

func tagsToInstAPI(tags map[string]string) []*models.Tag {
	var res []*models.Tag

	for k, v := range tags {
		res = append(res, &models.Tag{
			Key:   k,
			Value: v,
		})
	}

	return res
}

func providerSettingsToInstaCloudProviders(
	dataCentres []*models.CassandraDataCentre,
	dataCentre *v1alpha1.CassandraDataCentre,
) {

	for i := range dataCentres {
		for j := range dataCentres[i].AWSSettings {
			if dataCentre.CloudProvider == "AWC_VPC" {
				dataCentres[i].AWSSettings[j].CustomVirtualNetworkID = dataCentre.CustomVirtualNetworkId
			}
			dataCentres[i].AWSSettings[j].EBSEncryptionKey = dataCentre.DiskEncryptionKey
		}
	}

	for i := range dataCentres {
		for j := range dataCentres[i].GCPSettings {
			if dataCentre.CloudProvider == "GCP" {
				dataCentres[i].GCPSettings[j].CustomVirtualNetworkID = dataCentre.CustomVirtualNetworkId
			}
		}
	}

	for i := range dataCentres {
		for j := range dataCentres[i].AzureSettings {
			dataCentres[i].AzureSettings[j].ResourceGroup = dataCentre.ResourceGroup
		}
	}
}

func ClusterStatusFromInstAPI(body []byte) (*v1alpha1.ClusterStatus, error) {
	var clusterStatusFromInst models.ClusterStatus
	err := json.Unmarshal(body, &clusterStatusFromInst)
	if err != nil {
		return nil, err
	}

	dataCentres := dataCentresFromInstAPI(clusterStatusFromInst.DataCentres)
	clusterStatus := &v1alpha1.ClusterStatus{
		ID:          clusterStatusFromInst.ID,
		Status:      clusterStatusFromInst.Status,
		DataCentres: dataCentres,
	}

	return clusterStatus, nil
}

func dataCentresFromInstAPI(instaDataCentres []*models.DataCentreStatus) []*v1alpha1.DataCentreStatus {
	var dataCentres []*v1alpha1.DataCentreStatus
	for _, dataCentre := range instaDataCentres {
		nodes := nodesFromInstAPI(dataCentre.Nodes)
		dataCentres = append(dataCentres, &v1alpha1.DataCentreStatus{
			ID:         dataCentre.ID,
			Status:     dataCentre.Status,
			Nodes:      nodes,
			NodeNumber: dataCentre.NumberOfNodes,
		})
	}

	return dataCentres
}

func nodesFromInstAPI(instaNodes []*models.Node) []*v1alpha1.Node {
	var nodes []*v1alpha1.Node
	for _, node := range instaNodes {
		nodes = append(nodes, &v1alpha1.Node{
			ID:             node.ID,
			Rack:           node.Rack,
			PublicAddress:  node.PublicAddress,
			PrivateAddress: node.PrivateAddress,
			Status:         node.Status,
			Size:           node.Size,
		})
	}
	return nodes
}
