package instaclustr

import (
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr/models"
)

func CassandraToInstAPI(cassandraSpec *v1alpha1.CassandraSpec) *models.CassandraCluster {
	cassandraInstTwoFactorDelete := cassandraTwoFactorDeleteToInstAPI(cassandraSpec.TwoFactorDelete)

	cassandra := &models.CassandraCluster{
		ClusterSpec: models.ClusterSpec{
			Name:                  cassandraSpec.ClusterName,
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
				CloudProvider:       dataCentre.Provider.Name,
				ProviderAccountName: dataCentre.Provider.AccountName,
			},
		}
		cassandraInstDC.ContinuousBackup = cassandraSpec.ContinuousBackup
		cassandraInstDC.ReplicationFactor = cassandraSpec.ReplicationFactor
		cassandraInstDC.PrivateIPBroadcastForDiscovery = cassandraSpec.PrivateIPBroadcastForDiscovery
		cassandraInstDC.ClientToClusterEncryption = cassandraSpec.ClientToClusterEncryption

		providerToInstaProviderSettings(cassandraInstDCs, dataCentre.Provider)

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

func providerToInstaProviderSettings(
	dataCentres []*models.CassandraDataCentre,
	provider *v1alpha1.CloudProvider,
) {

	for i := range dataCentres {
		for j := range dataCentres[i].AWSSettings {
			if provider.Name == "AWC_VPC" {
				dataCentres[i].AWSSettings[j].CustomVirtualNetworkID = provider.CustomVirtualNetworkId
			}
			dataCentres[i].AWSSettings[j].EBSEncryptionKey = provider.DiskEncryptionKey
		}
	}

	for i := range dataCentres {
		for j := range dataCentres[i].GCPSettings {
			if provider.Name == "GCP" {
				dataCentres[i].GCPSettings[j].CustomVirtualNetworkID = provider.CustomVirtualNetworkId
			}
		}
	}

	for i := range dataCentres {
		for j := range dataCentres[i].AzureSettings {
			dataCentres[i].AzureSettings[j].ResourceGroup = provider.ResourceGroup
		}
	}
}

func CassandraFromInstAPI(cassandraCluster *models.CassandraStatus) *v1alpha1.CassandraStatus {

	return &v1alpha1.CassandraStatus{
		ClusterID:       cassandraCluster.ClusterID,
		ClusterStatus:   cassandraCluster.ClusterStatus,
		OperationStatus: cassandraCluster.OperationStatus,
		CassandraDCStatus: &models.CassandraDCStatus{
			Status:                     cassandraCluster.ClusterStatus,
			ID:                         pgInstaCluster.ID,
			ClusterCertificateDownload: pgInstaCluster.ClusterCertificateDownload,
			DataCentres:                dataCentres,
		},
	}
}

func dataCentresFromInstAPI(instaDataCentres []*models.DataCentreStatusAPIv1) []*v1alpha1.DataCentreStatus {
	var dataCentres []*v1alpha1.DataCentreStatus
	for _, dataCentre := range instaDataCentres {
		nodes := nodesFromInstAPIv1(dataCentre.Nodes)
		dataCentres = append(dataCentres, &v1alpha1.DataCentreStatus{
			ID:        dataCentre.ID,
			Status:    dataCentre.Status,
			Nodes:     nodes,
			NodeCount: dataCentre.NodeCount,
		})
	}

	return dataCentres
}
