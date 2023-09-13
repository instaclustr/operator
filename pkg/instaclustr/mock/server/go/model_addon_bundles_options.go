/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

import (
	"errors"
	"os"
)

// AddonBundlesOptions - Options
type AddonBundlesOptions struct {

	// Accepts true/false. Enables Password Authentication and User Authorization.
	AuthnAuthz bool `json:"authnAuthz,omitempty"`

	// Accepts true/false. Enables Client ⇄ Node Encryption.
	ClientEncryption bool `json:"clientEncryption,omitempty"`

	// Enables broadcast of private IPs for auto-discovery.
	UsePrivateBroadcastRPCAddress bool `json:"usePrivateBroadcastRPCAddress,omitempty"`

	// Enables broadcast of private IPs for auto-discovery.
	LuceneEnabled bool `json:"luceneEnabled,omitempty"`

	// Enables commitlog backups and increases the frequency of the default snapshot backups. This option replaces options.backupEnabled.
	ContinuousBackupEnabled bool `json:"continuousBackupEnabled,omitempty"`

	// Default number of partitions to be assigned per topic.
	NumberPartitions int32 `json:"numberPartitions,omitempty"`

	// Enable to allow brokers to automatically create a topic, if it does not already exist, when messages are published to it.
	AutoCreateTopics bool `json:"autoCreateTopics,omitempty"`

	// Enable to allow topics to be deleted via the ic-kafka-topics tool.
	DeleteTopics bool `json:"deleteTopics,omitempty"`

	// Enable to provision dedicated zookeeper nodes.
	DedicatedZookeeper bool `json:"dedicatedZookeeper,omitempty"`

	// Dedicated Zookeeper node size.
	ZookeeperNodeSize string `json:"zookeeperNodeSize,omitempty"`

	// Number of Zookeeper nodes; must be 3 or 5.
	ZookeeperNodeCount int32 `json:"zookeeperNodeCount,omitempty"`

	// Advertised HostName is required for PrivateLink.
	AdvertisedHostName string `json:"advertisedHostName,omitempty"`

	// Create KRaft Mode with COLOCATED
	KraftMode string `json:"kraftMode,omitempty"`

	// Number of KRaft Controller Node count
	KraftControllerNodeCount string `json:"kraftControllerNodeCount,omitempty"`

	// Accepts true/false. Enables Integration of the Karapace REST Proxy with the local Karapace Schema Registry
	IntegrateRestProxyWithSchemaRegistry bool `json:"integrateRestProxyWithSchemaRegistry,omitempty"`

	// Accepts true/false. Integrates the REST proxy with the Schema registry attached to this cluster. Requires 'integrateRestProxyWithSchemaRegistry' to be 'true'
	UseLocalSchemaRegistry bool `json:"useLocalSchemaRegistry,omitempty"`

	// URL of the Kafka schema registry to integrate with. Requires 'integrateRestProxyWithSchemaRegistry' to be 'true' and useLocalSchemaRegistry to be 'false'
	SchemaRegistryServerUrl string `json:"schemaRegistryServerUrl,omitempty"`

	// Username to use when connecting to the Kafka schema registry. Requires 'integrateRestProxyWithSchemaRegistry' to be 'true' and useLocalSchemaRegistry to be 'false'
	SchemaRegistryUsername string `json:"schemaRegistryUsername,omitempty"`

	// Password to use when connecting to the Kafka schema registry. Requires 'integrateRestProxyWithSchemaRegistry' to be 'true' and useLocalSchemaRegistry to be 'false'
	SchemaRegistryPassword string `json:"schemaRegistryPassword,omitempty"`

	// GUID of the Instaclustr managed Kafka Cluster Id you wish to connect to. Must be in the same Instaclustr account.
	TargetKafkaClusterId string `json:"targetKafkaClusterId,omitempty"`

	// Only required if targeting an Instaclustr managed cluster. If the target Kafka cluster does not have Private Network add-on, the endpoint only accepts vpcType of VPC_PEERED;  If the target Kafka cluster has the Private Network add-on, the endpoint accepts vpcTypes of VPC_PEERED and KAFKA_VPC.
	VpcType string `json:"vpcType,omitempty"`

	// Base64 encoded version of the TLS trust store (in JKS format) used to connect to your Kafka Cluster. Only required if connecting to a Non-Instaclustr managed Kafka Cluster with TLS enabled.
	Truststore *os.File `json:"truststore,omitempty"`

	// <b>AWS ONLY</b>. Access information for the S3 bucket where you will store your custom connectors.
	AwsAccessKeyId string `json:"aws.access.key.id,omitempty"`

	// <b>AWS ONLY</b>. Access information for the S3 bucket where you will store your custom connectors.
	AwsSecretAccessKey string `json:"aws.secret.access.key,omitempty"`

	// <b>AWS ONLY</b>. Access information for the S3 bucket where you will store your custom connectors.
	AwsS3RoleArn string `json:"aws.s3.role.arn,omitempty"`

	// <b>AWS ONLY</b>. Access information for the S3 bucket where you will store your custom connectors.
	S3BucketName string `json:"s3.bucket.name,omitempty"`

	// <b>AZURE ONLY</b>. Access information for the Azure Storage container where you will store your custom connectors.
	AzureStorageAccountName string `json:"azure.storage.account.name,omitempty"`

	// <b>AZURE ONLY</b>. Access information for the Azure Storage container where you will store your custom connectors.
	AzureStorageAccountKey string `json:"azure.storage.account.key,omitempty"`

	// <b>AZURE ONLY</b>. Access information for the Azure Storage container where you will store your custom connectors.
	AzureStorageContainerName string `json:"azure.storage.container.name,omitempty"`

	// This is an optional boolean parameter used to scale the load of the data node by having 3 dedicated master nodes for a particular cluster. We recommend selecting this to true for clusters with 9 and above nodes.
	DedicatedMasterNodes bool `json:"dedicatedMasterNodes,omitempty"`

	// Only to be included if the if dedicatedMasterNodes is set to true. See Available cluster node sizes in <a href='#operation/extendedProvisionRequestHandler'>Provisioning API</a>
	MasterNodeSize string `json:"masterNodeSize,omitempty"`

	// This is an optional string parameter to used to define the OpenSearch Dashboards node size. Only to be included if wish to add an OpenSearch Dashboards node to the cluster. See Available cluster node sizes in <a href='#operation/extendedProvisionRequestHandler'>Provisioning API</a>
	OpenSearchDashboardsNodeSize string `json:"openSearchDashboardsNodeSize,omitempty"`

	// Enables Index Management Plugin. This helps automate recurring index management activities.
	IndexManagementPlugin bool `json:"indexManagementPlugin,omitempty"`

	// Enables the Anomaly Detection Plugin. This allows for the detection of unusual or anomalous data points in your time-series data. Supported on OpenSearch 2.x for versions >= 2.4.0 and 1.x for versions >= 1.3.7
	AnomalyDetectionPlugin bool `json:"anomalyDetectionPlugin,omitempty"`

	// Enables the SQL Plugin. This allows you to query you indices with SQL as an alternative to the OpenSearch Query DSL. Supported on OpenSearch 2.x for versions >= 2.4.0 and 1.x for versions >= 1.3.7
	SqlPlugin bool `json:"sqlPlugin,omitempty"`

	// Enables the Anomaly Detection Plugin. This allows you to submit queries to be completed asynchronously. Supported on OpenSearch 2.x for versions >= 2.4.0 and 1.x for versions 1.3.7
	AsynchronousSearchPlugin bool `json:"asynchronousSearchPlugin,omitempty"`

	// Enables Security Plugin. This option gives an extra layer of security to the cluster. This will automatically enable internode encryption.
	SecurityPlugin bool `json:"securityPlugin,omitempty"`

	// Accepts true/false. Enables Password Authentication and User Authorization.
	PasswordAuth bool `json:"passwordAuth,omitempty"`

	// The number of Redis Master Nodes to be created.
	MasterNodes int32 `json:"masterNodes,omitempty"`

	// The number of Redis Replica Nodes to be created.
	ReplicaNodes int32 `json:"replicaNodes,omitempty"`

	// Require clients to provide credentials — a username & API Key — to connect to the Spark Jobserver.
	PasswordAuthentication bool `json:"passwordAuthentication,omitempty"`

	// The number of PostgreSQL server nodes created. One of which will be the primary while the rest will be replicas
	PostgresqlNodeCount int32 `json:"postgresqlNodeCount,omitempty"`

	// Default <a href=\"https://www.instaclustr.com/support/documentation/postgresql/options/replication-mode\">replication mode</a> for PostgreSQL transactions.
	ReplicationMode string `json:"replicationMode,omitempty"`

	// Accepts true/false. Enables <a href=\"https://www.instaclustr.com/support/documentation/postgresql/options/synchronous-mode-strict/\">synchronous mode strict</a> for clusters with SYNCHRONOUS replication mode.
	SynchronousModeStrict bool `json:"synchronousModeStrict,omitempty"`

	// PgBouncer pooling mode
	PoolMode string `json:"poolMode,omitempty"`

	// Use Advanced Visibility.
	UseAdvancedVisibility bool `json:"useAdvancedVisibility,omitempty"`

	// UUID of a Data Centre of an Instaclustr managed Cassandra Cluster, to be used by Cadence. Must be in the same Instaclustr account.
	TargetCassandraCdcId string `json:"targetCassandraCdcId"`

	// Target Cassandra VPC Type
	TargetCassandraVpcType string `json:"targetCassandraVpcType,omitempty"`

	// UUID of a Data Centre of an Instaclustr managed Kafka Cluster, to be used by Cadence. Must be in the same Instaclustr account.
	TargetKafkaCdcId string `json:"targetKafkaCdcId,omitempty"`

	// Target Kafka VPC Type
	TargetKafkaVpcType string `json:"targetKafkaVpcType,omitempty"`

	// UUID of a Data Centre of an Instaclustr managed OpenSearch Cluster, to be used by Cadence. Must be in the same Instaclustr account.
	TargetOpenSearchCdcId string `json:"targetOpenSearchCdcId,omitempty"`

	// Target OpenSearch VPC Type
	TargetOpenSearchVpcType string `json:"targetOpenSearchVpcType,omitempty"`

	// Use Cadence Archival
	EnableArchival bool `json:"enableArchival,omitempty"`

	// URI of S3 resource for Cadence Archival e.g. 's3://my-bucket'
	ArchivalS3Uri string `json:"archivalS3Uri,omitempty"`

	// Region of S3 resource for Cadence Archival e.g. 'us-east-1'
	ArchivalS3Region string `json:"archivalS3Region,omitempty"`

	// Provisioning Type used for the Cadence cluster. If SHARED is used, then the dependant clusters will be set to the Shared Infrastructure clusters.
	ProvisioningType string `json:"provisioningType,omitempty"`
}

// AssertAddonBundlesOptionsRequired checks if the required fields are not zero-ed
func AssertAddonBundlesOptionsRequired(obj AddonBundlesOptions) error {
	elements := map[string]interface{}{
		"targetCassandraCdcId": obj.TargetCassandraCdcId,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertAddonBundlesOptionsConstraints checks if the values respects the defined constraints
func AssertAddonBundlesOptionsConstraints(obj AddonBundlesOptions) error {
	if obj.MasterNodes < 3 {
		return &ParsingError{Err: errors.New(errMsgMinValueConstraint)}
	}
	if obj.MasterNodes > 50 {
		return &ParsingError{Err: errors.New(errMsgMaxValueConstraint)}
	}
	if obj.ReplicaNodes < 0 {
		return &ParsingError{Err: errors.New(errMsgMinValueConstraint)}
	}
	if obj.ReplicaNodes > 50 {
		return &ParsingError{Err: errors.New(errMsgMaxValueConstraint)}
	}
	if obj.PostgresqlNodeCount < 1 {
		return &ParsingError{Err: errors.New(errMsgMinValueConstraint)}
	}
	if obj.PostgresqlNodeCount > 5 {
		return &ParsingError{Err: errors.New(errMsgMaxValueConstraint)}
	}
	return nil
}