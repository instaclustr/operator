/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package instaclustr

import "time"

const (
	DefaultTimeout  = time.Second * 60
	OperatorVersion = "k8s v0.1.1"
)

// constants for API v2
const (
	ClustersEndpoint                       = "/cluster-management/v2/data-sources/clusters/v2/"
	CassandraEndpoint                      = "/cluster-management/v2/resources/applications/cassandra/clusters/v2/"
	OpenSearchEndpoint                     = "/cluster-management/v2/resources/applications/opensearch/clusters/v2/"
	KafkaEndpoint                          = "/cluster-management/v2/resources/applications/kafka/clusters/v2/"
	KafkaConnectEndpoint                   = "/cluster-management/v2/resources/applications/kafka-connect/clusters/v2/"
	KafkaMirrorEndpoint                    = "/cluster-management/v2/resources/applications/kafka_connect/mirrors/v2/"
	KafkaTopicEndpoint                     = "/cluster-management/v2/resources/applications/kafka/topics/v2/"
	KafkaTopicConfigsUpdateEndpoint        = "/cluster-management/v2/operations/applications/kafka/topics/v2/%s/modify-configs/v2"
	ZookeeperEndpoint                      = "/cluster-management/v2/resources/applications/zookeeper/clusters/v2/"
	AWSPeeringEndpoint                     = "/cluster-management/v2/resources/providers/aws/vpc-peers/v2/"
	AzurePeeringEndpoint                   = "/cluster-management/v2/resources/providers/azure/vnet-peers/v2/"
	ClusterNetworkFirewallRuleEndpoint     = "/cluster-management/v2/resources/network-firewall-rules/v2/"
	AWSSecurityGroupFirewallRuleEndpoint   = "/cluster-management/v2/resources/providers/aws/security-group-firewall-rules/v2/"
	GCPPeeringEndpoint                     = "/cluster-management/v2/resources/providers/gcp/vpc-peers/v2/"
	KafkaUserEndpoint                      = "/cluster-management/v2/resources/applications/kafka/users/v2/"
	KafkauserCertificatesEndpoint          = "/cluster-management/v2/resources/applications/kafka/user-certificates/v2/"
	KafkaUserCertificatesRenewEndpoint     = "/cluster-management/v2/operations/applications/kafka/user-certificates/renew/v2/"
	KafkaACLEndpoint                       = "/cluster-management/v2/resources/applications/kafka/acls/v2/"
	CadenceEndpoint                        = "/cluster-management/v2/resources/applications/cadence/clusters/v2/"
	RedisEndpoint                          = "/cluster-management/v2/resources/applications/redis/clusters/v2/"
	RedisUserEndpoint                      = "/cluster-management/v2/resources/applications/redis/users/v2/"
	PGSQLEndpoint                          = "/cluster-management/v2/resources/applications/postgresql/clusters/v2/"
	PGSQLConfigEndpoint                    = "%s/cluster-management/v2/data-sources/postgresql_cluster/%s/configurations"
	PGSQLConfigManagementEndpoint          = "%s/cluster-management/v2/resources/applications/postgresql/configurations/v2/"
	PGSQLUpdateDefaultUserPasswordEndpoint = "%s/cluster-management/v2/operations/applications/postgresql/clusters/v2/%s/update-default-user-password"
	NodeReloadEndpoint                     = "%s/cluster-management/v2/operations/applications/postgresql/nodes/v2/%s/reload"
	AWSEncryptionKeyEndpoint               = "/cluster-management/v2/resources/providers/aws/encryption-keys/v2/"
	ListAppsVersionsEndpoint               = "%s/cluster-management/v2/data-sources/applications/%s/versions/v2/"
	ClusterSettingsEndpoint                = "%s/cluster-management/v2/operations/clusters/v2/%s/change-settings/v2"
	AWSEndpointServicePrincipalEndpoint    = "/cluster-management/v2/resources/aws-endpoint-service-principals/v2/"
)

// constants for API v1
const (
	BackupsEndpoint = "/backups"

	// ClustersEndpoint is used for GET, DELETE and UPDATE clusters
	ClustersEndpointV1 = "/provisioning/v1/"

	APIv1RestoreEndpoint = "%s/provisioning/v1/%s/backups/restore"

	ExclusionWindowEndpoint        = "/v1/maintenance-events/exclusion-windows/"
	ExclusionWindowStatusEndpoint  = "/v1/maintenance-events/exclusion-windows?clusterId="
	MaintenanceEventEndpoint       = "/v1/maintenance-events/events/"
	MaintenanceEventStatusEndpoint = "/v1/maintenance-events/events?clusterId="
)

const (
	APIv1UserEndpoint   = "%s/provisioning/v1/%s/%s/users"
	RedisUserIDFmt      = "%s_%s"
	CassandraBundleUser = "apache_cassandra"
)
