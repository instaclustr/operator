/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package main

import (
	"log"
	"net/http"

	openapi "github.com/instaclustr/operator/pkg/instaclustr/mock/server/go"
)

func main() {
	log.Printf("Server started")

	AWSEncryptionKeyV2APIService := openapi.NewAWSEncryptionKeyV2APIService()
	AWSEncryptionKeyV2APIController := openapi.NewAWSEncryptionKeyV2APIController(AWSEncryptionKeyV2APIService)

	AWSEndpointServicePrincipalsV2APIService := openapi.NewAWSEndpointServicePrincipalsV2APIService()
	AWSEndpointServicePrincipalsV2APIController := openapi.NewAWSEndpointServicePrincipalsV2APIController(AWSEndpointServicePrincipalsV2APIService)

	AWSSecurityGroupFirewallRuleV2APIService := openapi.NewAWSSecurityGroupFirewallRuleV2APIService()
	AWSSecurityGroupFirewallRuleV2APIController := openapi.NewAWSSecurityGroupFirewallRuleV2APIController(AWSSecurityGroupFirewallRuleV2APIService)

	AWSVPCPeerV2APIService := openapi.NewAWSVPCPeerV2APIService()
	AWSVPCPeerV2APIController := openapi.NewAWSVPCPeerV2APIController(AWSVPCPeerV2APIService)

	AccountClusterListV2APIService := openapi.NewAccountClusterListV2APIService()
	AccountClusterListV2APIController := openapi.NewAccountClusterListV2APIController(AccountClusterListV2APIService)

	ApacheCassandraProvisioningV2APIService := openapi.NewApacheCassandraProvisioningV2APIService()
	ApacheCassandraProvisioningV2APIController := openapi.NewApacheCassandraProvisioningV2APIController(ApacheCassandraProvisioningV2APIService)

	ApacheKafkaACLV2APIService := openapi.NewApacheKafkaACLV2APIService()
	ApacheKafkaACLV2APIController := openapi.NewApacheKafkaACLV2APIController(ApacheKafkaACLV2APIService)

	ApacheKafkaConnectMirrorV2APIService := openapi.NewApacheKafkaConnectMirrorV2APIService()
	ApacheKafkaConnectMirrorV2APIController := openapi.NewApacheKafkaConnectMirrorV2APIController(ApacheKafkaConnectMirrorV2APIService)

	ApacheKafkaConnectProvisioningV2APIService := openapi.NewApacheKafkaConnectProvisioningV2APIService()
	ApacheKafkaConnectProvisioningV2APIController := openapi.NewApacheKafkaConnectProvisioningV2APIController(ApacheKafkaConnectProvisioningV2APIService)

	ApacheKafkaProvisioningV2APIService := openapi.NewApacheKafkaProvisioningV2APIService()
	ApacheKafkaProvisioningV2APIController := openapi.NewApacheKafkaProvisioningV2APIController(ApacheKafkaProvisioningV2APIService)

	ApacheKafkaRestProxyUserAPIService := openapi.NewApacheKafkaRestProxyUserAPIService()
	ApacheKafkaRestProxyUserAPIController := openapi.NewApacheKafkaRestProxyUserAPIController(ApacheKafkaRestProxyUserAPIService)

	ApacheKafkaSchemaRegistryUserAPIService := openapi.NewApacheKafkaSchemaRegistryUserAPIService()
	ApacheKafkaSchemaRegistryUserAPIController := openapi.NewApacheKafkaSchemaRegistryUserAPIController(ApacheKafkaSchemaRegistryUserAPIService)

	ApacheKafkaTopicV2APIService := openapi.NewApacheKafkaTopicV2APIService()
	ApacheKafkaTopicV2APIController := openapi.NewApacheKafkaTopicV2APIController(ApacheKafkaTopicV2APIService)

	ApacheKafkaUserAPIService := openapi.NewApacheKafkaUserAPIService()
	ApacheKafkaUserAPIController := openapi.NewApacheKafkaUserAPIController(ApacheKafkaUserAPIService)

	ApacheZookeeperProvisioningV2APIService := openapi.NewApacheZookeeperProvisioningV2APIService()
	ApacheZookeeperProvisioningV2APIController := openapi.NewApacheZookeeperProvisioningV2APIController(ApacheZookeeperProvisioningV2APIService)

	AzureVnetPeerV2APIService := openapi.NewAzureVnetPeerV2APIService()
	AzureVnetPeerV2APIController := openapi.NewAzureVnetPeerV2APIController(AzureVnetPeerV2APIService)

	CadenceProvisioningV2APIService := openapi.NewCadenceProvisioningV2APIService()
	CadenceProvisioningV2APIController := openapi.NewCadenceProvisioningV2APIController(CadenceProvisioningV2APIService)

	ClusterExclusionWindowV2APIService := openapi.NewClusterExclusionWindowV2APIService()
	ClusterExclusionWindowV2APIController := openapi.NewClusterExclusionWindowV2APIController(ClusterExclusionWindowV2APIService)

	ClusterMaintenanceEventsV2APIService := openapi.NewClusterMaintenanceEventsV2APIService()
	ClusterMaintenanceEventsV2APIController := openapi.NewClusterMaintenanceEventsV2APIController(ClusterMaintenanceEventsV2APIService)

	ClusterNetworkFirewallRuleV2APIService := openapi.NewClusterNetworkFirewallRuleV2APIService()
	ClusterNetworkFirewallRuleV2APIController := openapi.NewClusterNetworkFirewallRuleV2APIController(ClusterNetworkFirewallRuleV2APIService)

	ClusterSettingsV2APIService := openapi.NewClusterSettingsV2APIService()
	ClusterSettingsV2APIController := openapi.NewClusterSettingsV2APIController(ClusterSettingsV2APIService)

	GCPVPCPeerV2APIService := openapi.NewGCPVPCPeerV2APIService()
	GCPVPCPeerV2APIController := openapi.NewGCPVPCPeerV2APIController(GCPVPCPeerV2APIService)

	KarapaceRestProxyUserAPIService := openapi.NewKarapaceRestProxyUserAPIService()
	KarapaceRestProxyUserAPIController := openapi.NewKarapaceRestProxyUserAPIController(KarapaceRestProxyUserAPIService)

	KarapaceSchemaRegistryUserAPIService := openapi.NewKarapaceSchemaRegistryUserAPIService()
	KarapaceSchemaRegistryUserAPIController := openapi.NewKarapaceSchemaRegistryUserAPIController(KarapaceSchemaRegistryUserAPIService)

	OpenSearchEgressRulesV2APIService := openapi.NewOpenSearchEgressRulesV2APIService()
	OpenSearchEgressRulesV2APIController := openapi.NewOpenSearchEgressRulesV2APIController(OpenSearchEgressRulesV2APIService)

	OpenSearchProvisioningV2APIService := openapi.NewOpenSearchProvisioningV2APIService()
	OpenSearchProvisioningV2APIController := openapi.NewOpenSearchProvisioningV2APIController(OpenSearchProvisioningV2APIService)

	PostgreSQLConfigurationV2APIService := openapi.NewPostgreSQLConfigurationV2APIService()
	PostgreSQLConfigurationV2APIController := openapi.NewPostgreSQLConfigurationV2APIController(PostgreSQLConfigurationV2APIService)

	PostgreSQLProvisioningV2APIService := openapi.NewPostgreSQLProvisioningV2APIService()
	PostgreSQLProvisioningV2APIController := openapi.NewPostgreSQLProvisioningV2APIController(PostgreSQLProvisioningV2APIService)

	PostgreSQLReloadOperationV2APIService := openapi.NewPostgreSQLReloadOperationV2APIService()
	PostgreSQLReloadOperationV2APIController := openapi.NewPostgreSQLReloadOperationV2APIController(PostgreSQLReloadOperationV2APIService)

	PostgreSQLUserV2APIService := openapi.NewPostgreSQLUserV2APIService()
	PostgreSQLUserV2APIController := openapi.NewPostgreSQLUserV2APIController(PostgreSQLUserV2APIService)

	RedisProvisioningV2APIService := openapi.NewRedisProvisioningV2APIService()
	RedisProvisioningV2APIController := openapi.NewRedisProvisioningV2APIController(RedisProvisioningV2APIService)

	RedisUserV2APIService := openapi.NewRedisUserV2APIService()
	RedisUserV2APIController := openapi.NewRedisUserV2APIController(RedisUserV2APIService)

	ResizeOperationsV2APIService := openapi.NewResizeOperationsV2APIService()
	ResizeOperationsV2APIController := openapi.NewResizeOperationsV2APIController(ResizeOperationsV2APIService)

	BundleUserAPIService := openapi.NewBundleUserAPIService()
	BundleUserAPIController := openapi.NewBundleUserAPIController(BundleUserAPIService)

	router := openapi.NewRouter(BundleUserAPIController, AWSEncryptionKeyV2APIController, AWSEndpointServicePrincipalsV2APIController, AWSSecurityGroupFirewallRuleV2APIController, AWSVPCPeerV2APIController, AccountClusterListV2APIController, ApacheCassandraProvisioningV2APIController, ApacheKafkaACLV2APIController, ApacheKafkaConnectMirrorV2APIController, ApacheKafkaConnectProvisioningV2APIController, ApacheKafkaProvisioningV2APIController, ApacheKafkaRestProxyUserAPIController, ApacheKafkaSchemaRegistryUserAPIController, ApacheKafkaTopicV2APIController, ApacheKafkaUserAPIController, ApacheZookeeperProvisioningV2APIController, AzureVnetPeerV2APIController, CadenceProvisioningV2APIController, ClusterExclusionWindowV2APIController, ClusterMaintenanceEventsV2APIController, ClusterNetworkFirewallRuleV2APIController, ClusterSettingsV2APIController, GCPVPCPeerV2APIController, KarapaceRestProxyUserAPIController, KarapaceSchemaRegistryUserAPIController, OpenSearchEgressRulesV2APIController, OpenSearchProvisioningV2APIController, PostgreSQLConfigurationV2APIController, PostgreSQLProvisioningV2APIController, PostgreSQLReloadOperationV2APIController, PostgreSQLUserV2APIController, RedisProvisioningV2APIController, RedisUserV2APIController, ResizeOperationsV2APIController)

	log.Fatal(http.ListenAndServe(":8082", router))
}
