package convertors

import (
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/models"
	"github.com/instaclustr/operator/pkg/models"
)

func OpenSearchToInstAPI(openSearchSpec *v1alpha1.OpenSearchSpec) *modelsv1.OpenSearchCluster {
	openSearchBundles := openSearchBundlesToInstAPI(openSearchSpec.DataCentres[0], openSearchSpec.Version)

	openSearchInstProvider := openSearchProviderToInstAPI(openSearchSpec.DataCentres[0])

	openSearchInstTwoFactorDelete := twoFactorDeleteToInstAPI(openSearchSpec.TwoFactorDelete)

	openSearchRackAllocation := &models.RackAllocation{
		NodesPerRack:  openSearchSpec.DataCentres[0].NodesNumber,
		NumberOfRacks: openSearchSpec.DataCentres[0].RacksNumber,
	}

	return &modelsv1.OpenSearchCluster{
		Cluster: models.Cluster{
			ClusterName:           openSearchSpec.Name,
			NodeSize:              openSearchSpec.DataCentres[0].NodeSize,
			PrivateNetworkCluster: openSearchSpec.PrivateNetworkCluster,
			SLATier:               openSearchSpec.SLATier,
			Provider:              openSearchInstProvider,
			TwoFactorDelete:       openSearchInstTwoFactorDelete,
			DataCentre:            openSearchSpec.DataCentres[0].Region,
			DataCentreCustomName:  openSearchSpec.DataCentres[0].Name,
			ClusterNetwork:        openSearchSpec.DataCentres[0].Network,
			RackAllocation:        openSearchRackAllocation,
		},
		Bundles: openSearchBundles,
	}
}

func openSearchBundlesToInstAPI(
	dataCentre *v1alpha1.OpenSearchDataCentre,
	version string,
) []*modelsv1.OpenSearchBundle {
	var openSearchBundles []*modelsv1.OpenSearchBundle
	openSearchBundle := &modelsv1.OpenSearchBundle{
		Bundle: models.Bundle{
			Bundle:  modelsv1.OpenSearch,
			Version: version,
		},
		Options: &modelsv1.OpenSearchBundleOptions{
			DedicatedMasterNodes:         dataCentre.DedicatedMasterNodes,
			MasterNodeSize:               dataCentre.MasterNodeSize,
			OpenSearchDashboardsNodeSize: dataCentre.OpenSearchDashboardsNodeSize,
			IndexManagementPlugin:        dataCentre.IndexManagementPlugin,
		},
	}
	openSearchBundles = append(openSearchBundles, openSearchBundle)

	return openSearchBundles
}

func openSearchProviderToInstAPI(dataCentre *v1alpha1.OpenSearchDataCentre) *models.ClusterProvider {
	var instCustomVirtualNetworkId string
	var instResourceGroup string
	var insDiskEncryptionKey string
	if len(dataCentre.CloudProviderSettings) > 0 {
		instCustomVirtualNetworkId = dataCentre.CloudProviderSettings[0].CustomVirtualNetworkID
		instResourceGroup = dataCentre.CloudProviderSettings[0].ResourceGroup
		insDiskEncryptionKey = dataCentre.CloudProviderSettings[0].DiskEncryptionKey
	}

	return &models.ClusterProvider{
		Name:                   dataCentre.CloudProvider,
		AccountName:            dataCentre.ProviderAccountName,
		Tags:                   dataCentre.Tags,
		CustomVirtualNetworkId: instCustomVirtualNetworkId,
		ResourceGroup:          instResourceGroup,
		DiskEncryptionKey:      insDiskEncryptionKey,
	}
}
