package convertors

import (
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
)

type CreateKafka struct {
	PCIComplianceMode                 bool                      `json:"pciComplianceMode"`
	SchemaRegistry                    []*SchemaRegistry         `json:"schemaRegistry,omitempty"`
	DefaultReplicationFactor          int32                     `json:"defaultReplicationFactor"`
	DefaultNumberOfPartitions         int32                     `json:"defaultNumberOfPartitions"`
	RestProxy                         []*RestProxy              `json:"restProxy,omitempty"`
	TwoFactorDelete                   []*models.TwoFactorDelete `json:"twoFactorDelete,omitempty"`
	AllowDeleteTopics                 bool                      `json:"allowDeleteTopics"`
	AutoCreateTopics                  bool                      `json:"autoCreateTopics"`
	ClientToClusterEncryption         bool                      `json:"clientToClusterEncryption"`
	KafkaDataCentre                   []*models.DataCentre      `json:"dataCentres"`
	DedicatedZookeeper                []*DedicatedZookeeper     `json:"dedicatedZookeeper,omitempty"`
	PrivateNetworkCluster             bool                      `json:"privateNetworkCluster"`
	KafkaVersion                      string                    `json:"kafkaVersion"`
	Name                              string                    `json:"name"`
	SLATier                           string                    `json:"slaTier"`
	ClientBrokerAuthWithMTLS          bool                      `json:"clientBrokerAuthWithMtls,omitempty"`
	ClientAuthBrokerWithoutEncryption bool                      `json:"clientAuthBrokerWithoutEncryption,omitempty"`
	ClientAuthBrokerWithEncryption    bool                      `json:"clientAuthBrokerWithEncryption,omitempty"`
	KarapaceRestProxy                 []*KarapaceRestProxy      `json:"karapaceRestProxy,omitempty"`
	KarapaceSchemaRegistry            []*KarapaceSchemaRegistry `json:"karapaceSchemaRegistry,omitempty"`
}

type SchemaRegistry struct {
	Version string `json:"version"`
}

type RestProxy struct {
	IntegrateRestProxyWithSchemaRegistry bool   `json:"integrateRestProxyWithSchemaRegistry"`
	UseLocalSchemaRegistry               bool   `json:"useLocalSchemaRegistry,omitempty"`
	SchemaRegistryServerURL              string `json:"schemaRegistryServerUrl,omitempty"`
	SchemaRegistryUsername               string `json:"schemaRegistryUsername,omitempty"`
	SchemaRegistryPassword               string `json:"schemaRegistryPassword,omitempty"`
	Version                              string `json:"version"`
}

type KarapaceRestProxy struct {
	IntegrateRestProxyWithSchemaRegistry bool   `json:"integrateRestProxyWithSchemaRegistry"`
	Version                              string `json:"version"`
}

type KarapaceSchemaRegistry struct {
	Version string `json:"version"`
}

type DedicatedZookeeper struct {
	ZookeeperNodeSize  string `json:"zookeeperNodeSize"`
	ZookeeperNodeCount int32  `json:"zookeeperNodeCount"`
}

func KafkaToInstAPI(k v1alpha1.KafkaSpec) CreateKafka {
	return CreateKafka{
		SchemaRegistry:                    schemaRegistryToInstAPI(k.SchemaRegistry),
		RestProxy:                         restProxyToInstAPI(k.RestProxy),
		PCIComplianceMode:                 k.PCICompliance,
		DefaultReplicationFactor:          k.ReplicationFactorNumber,
		DefaultNumberOfPartitions:         k.PartitionsNumber,
		TwoFactorDelete:                   twoFactorDeleteToInstAPI(k.TwoFactorDelete),
		AllowDeleteTopics:                 k.AllowDeleteTopics,
		AutoCreateTopics:                  k.AutoCreateTopics,
		ClientToClusterEncryption:         k.ClientToClusterEncryption,
		DedicatedZookeeper:                dedicatedZookeeperToInstAPI(k.DedicatedZookeeper),
		PrivateNetworkCluster:             k.PrivateNetworkCluster,
		Name:                              k.Name,
		SLATier:                           k.SLATier,
		KafkaVersion:                      k.Version,
		KafkaDataCentre:                   kafkaDataCentresToInstAPI(k.DataCentres),
		ClientBrokerAuthWithMTLS:          k.ClientBrokerAuthWithMTLS,
		ClientAuthBrokerWithoutEncryption: k.ClientAuthBrokerWithoutEncryption,
		ClientAuthBrokerWithEncryption:    k.ClientAuthBrokerWithEncryption,
		KarapaceRestProxy:                 karapaceRestProxyToInstAPI(k.KarapaceRestProxy),
		KarapaceSchemaRegistry:            karapaceSchemaRegistryToInstAPI(k.KarapaceSchemaRegistry),
	}
}

func schemaRegistryToInstAPI(crdSchemas []*v1alpha1.SchemaRegistry) []*SchemaRegistry {
	if crdSchemas == nil {
		return nil
	}

	var instaSchemas []*SchemaRegistry
	for _, schema := range crdSchemas {
		instaSchemas = append(instaSchemas, &SchemaRegistry{
			Version: schema.Version,
		})
	}

	return instaSchemas
}

func karapaceSchemaRegistryToInstAPI(crdSchemas []*v1alpha1.KarapaceSchemaRegistry) []*KarapaceSchemaRegistry {
	if crdSchemas == nil {
		return nil
	}

	var instaSchemas []*KarapaceSchemaRegistry
	for _, schema := range crdSchemas {
		instaSchemas = append(instaSchemas, &KarapaceSchemaRegistry{
			Version: schema.Version,
		})
	}

	return instaSchemas
}

func restProxyToInstAPI(crdProxies []*v1alpha1.RestProxy) []*RestProxy {
	if crdProxies == nil {
		return nil
	}

	var instaRestProxies []*RestProxy
	for _, proxy := range crdProxies {
		instaRestProxies = append(instaRestProxies, &RestProxy{
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

func karapaceRestProxyToInstAPI(crdProxies []*v1alpha1.KarapaceRestProxy) []*KarapaceRestProxy {
	if crdProxies == nil {
		return nil
	}

	var instaRestProxies []*KarapaceRestProxy
	for _, proxy := range crdProxies {
		instaRestProxies = append(instaRestProxies, &KarapaceRestProxy{
			IntegrateRestProxyWithSchemaRegistry: proxy.IntegrateRestProxyWithSchemaRegistry,
			Version:                              proxy.Version,
		})
	}

	return instaRestProxies
}

func dedicatedZookeeperToInstAPI(crdZookeepers []*v1alpha1.DedicatedZookeeper) []*DedicatedZookeeper {
	if crdZookeepers == nil {
		return nil
	}

	var instaZookeepers []*DedicatedZookeeper
	for _, zookeeper := range crdZookeepers {
		instaZookeepers = append(instaZookeepers, &DedicatedZookeeper{
			ZookeeperNodeSize:  zookeeper.NodeSize,
			ZookeeperNodeCount: zookeeper.NodesNumber,
		})
	}

	return instaZookeepers
}

func kafkaDataCentresToInstAPI(crdDCs []*v1alpha1.KafkaDataCentre) []*models.DataCentre {
	if crdDCs == nil {
		return nil
	}

	var instaDCs []*models.DataCentre

	for _, crdDC := range crdDCs {
		instaDC := &models.DataCentre{
			Name:                crdDC.Name,
			Network:             crdDC.Network,
			NodeSize:            crdDC.NodeSize,
			NumberOfNodes:       crdDC.NodesNumber,
			Tags:                TagsToInstAPI(crdDC.Tags),
			CloudProvider:       crdDC.CloudProvider,
			Region:              crdDC.Region,
			ProviderAccountName: crdDC.ProviderAccountName,
		}

		allocateKafkaProviderSettingsToInstAPI(crdDC, instaDC)

		instaDCs = append(instaDCs, instaDC)
	}

	return instaDCs
}

func allocateKafkaProviderSettingsToInstAPI(crdDC *v1alpha1.KafkaDataCentre, instaDC *models.DataCentre) {
	for _, crdSetting := range crdDC.CloudProviderSettings {
		switch crdDC.CloudProvider {
		case models.AWSVPC:
			instaDC.AWSSettings = append(instaDC.AWSSettings, &models.AWSSetting{
				EBSEncryptionKey:       crdSetting.DiskEncryptionKey,
				CustomVirtualNetworkID: crdSetting.CustomVirtualNetworkID,
			})
		case models.GCP:
			instaDC.GCPSettings = append(instaDC.GCPSettings, &models.GCPSetting{
				CustomVirtualNetworkID: crdSetting.CustomVirtualNetworkID,
			})
		case models.AZURE, models.AZUREAZ:
			instaDC.AzureSettings = append(instaDC.AzureSettings, &models.AzureSetting{
				ResourceGroup: crdSetting.ResourceGroup,
			})
		}
	}

}

type updateKafkaInstAPI struct {
	DataCentre         []*models.DataCentre  `json:"dataCentres"`
	DedicatedZookeeper []*DedicatedZookeeper `json:"dedicatedZookeeper,omitempty"`
}

func KafkaToInstAPIUpdate(k *v1alpha1.KafkaSpec) *updateKafkaInstAPI {
	var newKafka updateKafkaInstAPI

	newKafka.DataCentre = kafkaDataCentresToInstAPI(k.DataCentres)
	newKafka.DedicatedZookeeper = dedicatedZookeeperToInstAPI(k.DedicatedZookeeper)

	return &newKafka
}
