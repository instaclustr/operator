package convertors

import (
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	models2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
)

func KafkaToInstAPI(k v1alpha1.KafkaSpec) models2.CreateKafka {
	return models2.CreateKafka{
		SchemaRegistry:            schemaRegistryToInstAPI(k.SchemaRegistry),
		RestProxy:                 restProxyToInstAPI(k.RestProxy),
		PCIComplianceMode:         k.PCICompliance,
		DefaultReplicationFactor:  k.ReplicationFactorNumber,
		DefaultNumberOfPartitions: k.PartitionsNumber,
		TwoFactorDelete:           twoFactorDeleteToInstAPI(k.TwoFactorDelete),
		AllowDeleteTopics:         k.AllowDeleteTopics,
		AutoCreateTopics:          k.AutoCreateTopics,
		ClientToClusterEncryption: k.ClientToClusterEncryption,
		DedicatedZookeeper:        dedicatedZookeeperToInstAPI(k.DedicatedZookeeper),
		PrivateNetworkCluster:     k.PrivateNetworkCluster,
		Name:                      k.Name,
		SLATier:                   k.SLATier,
		KafkaVersion:              k.Version,
		KafkaDataCentre:           kafkaDCToInstAPI(k.DataCentres),
	}
}

func schemaRegistryToInstAPI(crdSchemas []*v1alpha1.SchemaRegistry) []*models2.SchemaRegistry {
	if crdSchemas == nil {
		return nil
	}

	var instaSchemas []*models2.SchemaRegistry
	for _, schema := range crdSchemas {
		instaSchemas = append(instaSchemas, &models2.SchemaRegistry{
			Version: schema.Version,
		})
	}

	return instaSchemas
}

func restProxyToInstAPI(crdProxies []*v1alpha1.RestProxy) []*models2.RestProxy {
	if crdProxies == nil {
		return nil
	}

	var instaRestProxies []*models2.RestProxy
	for _, proxy := range crdProxies {
		instaRestProxies = append(instaRestProxies, &models2.RestProxy{
			IntegrateRestProxyWithSchemaRegistry: proxy.IntegrateRestProxyWithSchemaRegistry,
			UseLocalSchemaRegistry:               proxy.UseLocalSchemaRegistry,
			SchemaRegistryServerURL:              proxy.SchemaRegistryServerURL,
			SchemaRegistryUsername:               proxy.SchemaRegistryUsername,
			SchemaRegistryPassword:               proxy.SchemaRegistryPassword,
			Version:                              proxy.Version,
		})
	}

	return instaRestProxies
}

func dedicatedZookeeperToInstAPI(crdZookeepers []*v1alpha1.DedicatedZookeeper) []*models2.DedicatedZookeeper {
	if crdZookeepers == nil {
		return nil
	}

	var instaZookeepers []*models2.DedicatedZookeeper
	for _, zookeeper := range crdZookeepers {
		instaZookeepers = append(instaZookeepers, &models2.DedicatedZookeeper{
			ZookeeperNodeSize:  zookeeper.NodeSize,
			ZookeeperNodeCount: zookeeper.NodesNumber,
		})
	}

	return instaZookeepers
}

func kafkaDCToInstAPI(crdDCs []*v1alpha1.DataCentre) []models2.KafkaDataCentre {
	if crdDCs == nil {
		return nil
	}

	var instaDCs []models2.KafkaDataCentre

	for _, crdDC := range crdDCs {
		instaDC := models2.KafkaDataCentre{
			Name:                crdDC.Name,
			Network:             crdDC.Network,
			NodeSize:            crdDC.NodeSize,
			NumberOfNodes:       crdDC.NodesNumber,
			Tags:                tagsToInstAPI(crdDC.Tags),
			CloudProvider:       crdDC.CloudProvider,
			Region:              crdDC.Region,
			ProviderAccountName: crdDC.ProviderAccountName,
		}

		allocateProviderSettingsToInstAPI(crdDC, &instaDC)

		instaDCs = append(instaDCs, instaDC)
	}

	return instaDCs
}

func allocateProviderSettingsToInstAPI(crdDC *v1alpha1.DataCentre, instaDC *models2.KafkaDataCentre) {
	for _, crdSetting := range crdDC.CloudProviderSettings {
		switch crdDC.CloudProvider {
		case models2.AWSVPC:
			instaDC.AWSSettings = append(instaDC.AWSSettings, &models2.AWSSetting{
				EBSEncryptionKey:       crdSetting.DiskEncryptionKey,
				CustomVirtualNetworkID: crdSetting.CustomVirtualNetworkID,
			})
		case models2.GCP:
			instaDC.GCPSettings = append(instaDC.GCPSettings, &models2.GCPSetting{
				CustomVirtualNetworkID: crdSetting.CustomVirtualNetworkID,
			})
		case models2.AZURE, models2.AZUREAZ:
			instaDC.AzureSettings = append(instaDC.AzureSettings, &models2.AzureSetting{
				ResourceGroup: crdSetting.ResourceGroup,
			})
		}
	}

}
