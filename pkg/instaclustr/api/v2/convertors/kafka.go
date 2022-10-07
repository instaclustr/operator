package convertors

import (
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
)

func KafkaToInstAPI(k v1alpha1.KafkaSpec) modelsv2.CreateKafka {
	return modelsv2.CreateKafka{
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
		KafkaDataCentre:           dataCentresToInstAPI(k.DataCentres),
	}
}

func schemaRegistryToInstAPI(crdSchemas []*v1alpha1.SchemaRegistry) []*modelsv2.SchemaRegistry {
	if crdSchemas == nil {
		return nil
	}

	var instaSchemas []*modelsv2.SchemaRegistry
	for _, schema := range crdSchemas {
		instaSchemas = append(instaSchemas, &modelsv2.SchemaRegistry{
			Version: schema.Version,
		})
	}

	return instaSchemas
}

func restProxyToInstAPI(crdProxies []*v1alpha1.RestProxy) []*modelsv2.RestProxy {
	if crdProxies == nil {
		return nil
	}

	var instaRestProxies []*modelsv2.RestProxy
	for _, proxy := range crdProxies {
		instaRestProxies = append(instaRestProxies, &modelsv2.RestProxy{
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

func dedicatedZookeeperToInstAPI(crdZookeepers []*v1alpha1.DedicatedZookeeper) []*modelsv2.DedicatedZookeeper {
	if crdZookeepers == nil {
		return nil
	}

	var instaZookeepers []*modelsv2.DedicatedZookeeper
	for _, zookeeper := range crdZookeepers {
		instaZookeepers = append(instaZookeepers, &modelsv2.DedicatedZookeeper{
			ZookeeperNodeSize:  zookeeper.NodeSize,
			ZookeeperNodeCount: zookeeper.NodesNumber,
		})
	}

	return instaZookeepers
}
