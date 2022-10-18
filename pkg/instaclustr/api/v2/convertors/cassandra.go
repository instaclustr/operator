package convertors

import (
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
)

func CassandraToInstAPI(cassandraSpec *v1alpha1.CassandraSpec) *modelsv2.CassandraCluster {
	cassandraInstTwoFactorDelete := twoFactorDeleteToInstAPI(cassandraSpec.TwoFactorDelete)

	cassandra := &modelsv2.CassandraCluster{
		Cluster: modelsv2.Cluster{
			Name:                  cassandraSpec.Name,
			SLATier:               cassandraSpec.SLATier,
			PrivateNetworkCluster: cassandraSpec.PrivateNetworkCluster,
			PCIComplianceMode:     cassandraSpec.PCICompliance,
			TwoFactorDeletes:      cassandraInstTwoFactorDelete,
		},
		CassandraVersion:    cassandraSpec.Version,
		LuceneEnabled:       cassandraSpec.LuceneEnabled,
		PasswordAndUserAuth: cassandraSpec.PasswordAndUserAuth,
		Spark:               convertToInstaSpark(cassandraSpec.Spark),
	}

	var cassandraInstDCs []*modelsv2.CassandraDataCentre
	for _, dataCentre := range cassandraSpec.DataCentres {

		cassandraInstDC := &modelsv2.CassandraDataCentre{
			DataCentre: modelsv2.DataCentre{
				Name:                dataCentre.Name,
				Network:             dataCentre.Network,
				NodeSize:            dataCentre.NodeSize,
				NumberOfNodes:       dataCentre.NodesNumber,
				Region:              dataCentre.Region,
				Tags:                TagsToInstAPI(dataCentre.Tags),
				CloudProvider:       dataCentre.CloudProvider,
				ProviderAccountName: dataCentre.ProviderAccountName,
			},
			ContinuousBackup:               dataCentre.ContinuousBackup,
			PrivateIPBroadcastForDiscovery: dataCentre.PrivateIPBroadcastForDiscovery,
			ClientToClusterEncryption:      dataCentre.ClientToClusterEncryption,
			ReplicationFactor:              dataCentre.RacksNumber,
		}

		if len(dataCentre.CloudProviderSettings) > 0 {
			providerSettingsToInstaCloudProviders(dataCentre, cassandraInstDC)
		}

		cassandraInstDCs = append(cassandraInstDCs, cassandraInstDC)

	}
	cassandra.DataCentres = cassandraInstDCs

	return cassandra
}

func providerSettingsToInstaCloudProviders(dc *v1alpha1.CassandraDataCentre, cassandraDC *modelsv2.CassandraDataCentre) {
	for _, cp := range dc.CloudProviderSettings {
		switch dc.CloudProvider {
		case modelsv2.AWSVPC:
			cassandraDC.AWSSettings = append(cassandraDC.AWSSettings, &modelsv2.AWSSetting{
				EBSEncryptionKey:       "",
				CustomVirtualNetworkID: cp.CustomVirtualNetworkID,
			})
		case modelsv2.GCP:
			cassandraDC.GCPSettings = append(cassandraDC.GCPSettings, &modelsv2.GCPSetting{
				CustomVirtualNetworkID: cp.CustomVirtualNetworkID,
			})

		case modelsv2.AZURE, modelsv2.AZUREAZ:
			cassandraDC.AzureSettings = append(cassandraDC.AzureSettings, &modelsv2.AzureSetting{
				ResourceGroup: cp.ResourceGroup,
			})
		}
	}
}

func CompareCassandraDCs(k8sDCs []*v1alpha1.CassandraDataCentre,
	instaStatus *v1alpha1.ClusterStatus,
) *modelsv2.CassandraDCs {

	currentDCs := &modelsv2.CassandraDCs{
		DataCentres: cassandraSpecDCsToInstaDCs(k8sDCs),
	}

	if len(currentDCs.DataCentres) > len(instaStatus.DataCentres) {
		return currentDCs
	} else {
		for i := range currentDCs.DataCentres {
			if currentDCs.DataCentres[i].NumberOfNodes != instaStatus.DataCentres[i].NodeNumber {
				return currentDCs
			}
			for j := range instaStatus.DataCentres[i].Nodes {
				if currentDCs.DataCentres[i].NodeSize != instaStatus.DataCentres[i].Nodes[j].Size {
					return currentDCs
				}
			}
		}
	}

	return nil
}

func cassandraSpecDCsToInstaDCs(k8sDC []*v1alpha1.CassandraDataCentre) []*modelsv2.CassandraDataCentre {
	var cassandraInstDCs []*modelsv2.CassandraDataCentre
	for _, k8sCassandraDC := range k8sDC {
		dc := &modelsv2.CassandraDataCentre{
			DataCentre: modelsv2.DataCentre{
				Name:                k8sCassandraDC.Name,
				Network:             k8sCassandraDC.Network,
				NodeSize:            k8sCassandraDC.NodeSize,
				NumberOfNodes:       k8sCassandraDC.NodesNumber,
				Region:              k8sCassandraDC.Region,
				Tags:                TagsToInstAPI(k8sCassandraDC.Tags),
				CloudProvider:       k8sCassandraDC.CloudProvider,
				ProviderAccountName: k8sCassandraDC.ProviderAccountName,
			},
			ReplicationFactor:              k8sCassandraDC.RacksNumber,
			ContinuousBackup:               k8sCassandraDC.ContinuousBackup,
			PrivateIPBroadcastForDiscovery: k8sCassandraDC.PrivateIPBroadcastForDiscovery,
			ClientToClusterEncryption:      k8sCassandraDC.ClientToClusterEncryption,
		}

		if len(k8sCassandraDC.CloudProviderSettings) > 0 {
			providerSettingsToInstaCloudProviders(k8sCassandraDC, dc)
		}
		cassandraInstDCs = append(cassandraInstDCs, dc)
	}
	return cassandraInstDCs
}

func convertToInstaSpark(sparkVersionsFromSpec []*v1alpha1.Spark) []*modelsv2.Spark {
	var sparkVersions []*modelsv2.Spark
	for _, sp := range sparkVersionsFromSpec {
		sparkVersion := &modelsv2.Spark{
			Version: sp.Version,
		}
		sparkVersions = append(sparkVersions, sparkVersion)
	}

	return sparkVersions
}
