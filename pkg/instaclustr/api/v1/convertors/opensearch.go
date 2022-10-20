package convertors

import (
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/models"
)

func OpenSearchToInstAPI(openSearchSpec *v1alpha1.OpenSearchSpec) *modelsv1.OpenSearchCluster {
	dataCentresNumber := len(openSearchSpec.DataCentres)
	isSingleDC := checkSingleDCCluster(dataCentresNumber)

	openSearchBundles := openSearchBundlesToInstAPI(openSearchSpec.DataCentres[0], openSearchSpec.Version)

	openSearchInstProvider := openSearchProviderToInstAPI(openSearchSpec.DataCentres[0])

	openSearchInstTwoFactorDelete := twoFactorDeleteToInstAPI(openSearchSpec.TwoFactorDelete)

	openSearch := &modelsv1.OpenSearchCluster{
		Cluster: modelsv1.Cluster{
			ClusterName:           openSearchSpec.Name,
			NodeSize:              openSearchSpec.DataCentres[0].NodeSize,
			PrivateNetworkCluster: openSearchSpec.PrivateNetworkCluster,
			SLATier:               openSearchSpec.SLATier,
			Provider:              openSearchInstProvider,
			TwoFactorDelete:       openSearchInstTwoFactorDelete,
		},
		Bundles: openSearchBundles,
	}

	if isSingleDC {
		openSearchRackAllocation := &modelsv1.RackAllocation{
			NodesPerRack:  openSearchSpec.DataCentres[0].NodesNumber,
			NumberOfRacks: openSearchSpec.DataCentres[0].RacksNumber,
		}

		openSearch.DataCentre = openSearchSpec.DataCentres[0].Region
		openSearch.DataCentreCustomName = openSearchSpec.DataCentres[0].Name
		openSearch.NodeSize = openSearchSpec.DataCentres[0].NodeSize
		openSearch.ClusterNetwork = openSearchSpec.DataCentres[0].Network
		openSearch.RackAllocation = openSearchRackAllocation

		return openSearch
	}

	var openSearchInstDCs []*modelsv1.OpenSearchDataCentre
	for _, dataCentre := range openSearchSpec.DataCentres {
		openSearchBundles = openSearchBundlesToInstAPI(dataCentre, openSearchSpec.Version)

		openSearchInstProvider = openSearchProviderToInstAPI(dataCentre)

		openSearchRackAlloc := &modelsv1.RackAllocation{
			NodesPerRack:  dataCentre.NodesNumber,
			NumberOfRacks: dataCentre.RacksNumber,
		}

		openSearchInstDC := &modelsv1.OpenSearchDataCentre{
			DataCentre: modelsv1.DataCentre{
				Name:           dataCentre.Name,
				DataCentre:     dataCentre.Region,
				Network:        dataCentre.Network,
				Provider:       openSearchInstProvider,
				NodeSize:       dataCentre.NodeSize,
				RackAllocation: openSearchRackAlloc,
			},
			Bundles: openSearchBundles,
		}

		openSearchInstDCs = append(openSearchInstDCs, openSearchInstDC)
	}

	openSearch.DataCentres = openSearchInstDCs

	return openSearch
}

func openSearchBundlesToInstAPI(
	dataCentre *v1alpha1.OpenSearchDataCentre,
	version string,
) []*modelsv1.OpenSearchBundle {
	var openSearchBundles []*modelsv1.OpenSearchBundle
	openSearchBundle := &modelsv1.OpenSearchBundle{
		Bundle: modelsv1.Bundle{
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

func openSearchProviderToInstAPI(dataCentre *v1alpha1.OpenSearchDataCentre) *modelsv1.ClusterProvider {
	var instCustomVirtualNetworkId string
	var instResourceGroup string
	var insDiskEncryptionKey string
	if len(dataCentre.CloudProviderSettings) > 0 {
		instCustomVirtualNetworkId = dataCentre.CloudProviderSettings[0].CustomVirtualNetworkID
		instResourceGroup = dataCentre.CloudProviderSettings[0].ResourceGroup
		insDiskEncryptionKey = dataCentre.CloudProviderSettings[0].DiskEncryptionKey
	}

	return &modelsv1.ClusterProvider{
		Name:                   dataCentre.CloudProvider,
		AccountName:            dataCentre.ProviderAccountName,
		Tags:                   dataCentre.Tags,
		CustomVirtualNetworkId: instCustomVirtualNetworkId,
		ResourceGroup:          instResourceGroup,
		DiskEncryptionKey:      insDiskEncryptionKey,
	}
}
