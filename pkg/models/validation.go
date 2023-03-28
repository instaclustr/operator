package models

var (
	RedisVersions        = []string{"6.2.7", "7.0.5"}
	CassandraVersions    = []string{"4.0.4", "3.11.13"}
	SparkVersions        = []string{"2.3.2", "3.0.1"}
	CadenceVersions      = []string{"0.22.4", "0.24.0"}
	PostgreSQLVersions   = []string{"15.1.0", "14.6.0", "14.5.0", "13.9.0", "13.8.0"}
	PGBouncerVersions    = []string{"1.17.0"}
	KafkaVersions        = []string{"3.3.1", "3.0.2", "3.1.2", "2.8.2"}
	ZookeeperVersions    = []string{"3.6.3", "3.7.1"}
	KafkaConnectVersions = []string{"3.1.2", "3.0.2", "2.8.2", "2.7.1"}
	KafkaConnectVPCTypes = []string{"KAFKA_VPC", "VPC_PEERED", "SEPARATE_VPC"}
	OpenSearchVersions   = []string{"opensearch:1.3.7", "opensearch:2.2.1", "opensearch:1.3.7.ic1", "opensearch:2.2.1.ic1"}
	PoolModes            = []string{"TRANSACTION", "SESSION", "STATEMENT"}
	ReplicationModes     = []string{"ASYNCHRONOUS", "SYNCHRONOUS"}
	CloudProviders       = []string{"AWS_VPC", "GCP", "AZURE", "AZURE_AZ"}
	SLATiers             = []string{"PRODUCTION", "NON_PRODUCTION"}
	ClusterNameRegExp    = "^[a-zA-Z0-9][a-zA-Z0-9_-]{2,31}$"
	BundleTypes          = []string{"APACHE_ZOOKEEPER", "CADENCE", "CADENCE_GRPC",
		"CADENCE_WEB", "CASSANDRA", "CASSANDRA_CQL",
		"ELASTICSEARCH", "KAFKA", "KAFKA_CONNECT",
		"KAFKA_ENCRYPTION", "KAFKA_MTLS", "KAFKA_NO_ENCRYPTION",
		"KAFKA_REST_PROXY", "KAFKA_SCHEMA_REGISTRY", "KARAPACE_REST_PROXY",
		"KARAPACE_SCHEMA_REGISTRY", "OPENSEARCH", "OPENSEARCH_DASHBOARDS",
		"PGBOUNCER", "POSTGRESQL", "REDIS",
		"SEARCH_DASHBOARDS", "SECURE_APACHE_ZOOKEEPER", "SPARK",
		"SPARK_JOBSERVER", "SHOTOVER_PROXY"}
	PeerAWSAccountIDRegExp = "^[0-9]{12}$"
	PeerVPCIDRegExp        = "^vpc-[0-9a-f]{8}$"
	PeerSubnetsRegExp      = "^(10|172|192)\\.(25[0-5]|2[0-4][0-9]|[0-1]{1}[0-9]{2}|[1-9]{1}[0-9]{1}|[1-9]|0)\\.(25[0-5]|2[0-4][0-9]|[0-1]{1}[0-9]{2}|[1-9]{1}[0-9]{1}|[0-9])\\.(0)(\\/1[6-9]|\\/2[0-8])$"
	UUIDStringRegExp       = "^[0-9a-fA-F]{8}\\b-[0-9a-fA-F]{4}\\b-[0-9a-fA-F]{4}\\b-[0-9a-fA-F]{4}\\b-[0-9a-fA-F]{12}$"
	AWSRegions             = []string{"AF_SOUTH_1", "AP_EAST_1", "AP_NORTHEAST_1", "AP_NORTHEAST_2", "AP_SOUTHEAST_1",
		"AP_SOUTHEAST_2", "AP_SOUTH_1", "CA_CENTRAL_1", "CN_NORTHWEST_1", "CN_NORTH_1", "EU_CENTRAL_1", "EU_NORTH_1",
		"EU_SOUTH_1", "EU_WEST_1", "EU_WEST_2", "EU_WEST_3", "ME_SOUTH_1", "SA_EAST_1", "US_EAST_1", "US_EAST_2",
		"US_WEST_1", "US_WEST_2"}
	AzureRegions = []string{"AUSTRALIA_EAST", "CANADA_CENTRAL", "CENTRAL_US", "EAST_US", "EAST_US_2", "NORTH_EUROPE",
		"SOUTHEAST_ASIA", "SOUTH_CENTRAL_US", "WEST_EUROPE", "WEST_US_2"}
	GCPRegions = []string{"asia-east1", "asia-northeast1", "asia-south1", "asia-southeast1", "australia-southeast1",
		"europe-north1", "europe-west1", "europe-west2", "europe-west3", "europe-west4", "europe-west6",
		"northamerica-northeast1", "southamerica-east1", "us-central1", "us-east1", "us-east4", "us-west1", "us-west2"}
	DaysOfWeek          = []string{"MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY"}
	ISODateFormatRegExp = "^(?:[1-9]\\d{3}-(?:(?:0[1-9]|1[0-2])-(?:0[1-9]|1\\d|2[0-8])|(?:0[13-9]|1[0-2])-(?:29|30)|(?:0[13578]|1[02])-31)|(?:[1-9]\\d(?:0[48]|[2468][048]|[13579][26])|(?:[2468][048]|[13579][26])00)-02-29)T(?:[01]\\d|2[0-3]):[0-5]\\d:[0-5]\\d(?:Z|[+-][01]\\d:[0-5]\\d)$"
	ACLPermissionType   = []string{"ALLOW", "DENY"}
	ACLPatternType      = []string{"LITERAL", "PREFIXED"}
	ACLOperation        = []string{"ALL", "READ", "WRITE", "CREATE", "DELETE", "ALTER", "DESCRIBE", "CLUSTER_ACTION",
		"DESCRIBE_CONFIGS", "ALTER_CONFIGS", "IDEMPOTENT_WRITE"}
	ACLResourceType    = []string{"CLUSTER", "TOPIC", "GROUP", "DELEGATION_TOKEN", "TRANSACTIONAL_ID"}
	ACLUserPrefix      = "User:"
	ACLPrincipalRegExp = "^User:.*$"
	S3URIRegExp        = "^s3:\\/\\/[a-zA-Z0-9_-]+[^\\/]$"
	DependencyVPCs     = []string{"TARGET_VPC", "VPC_PEERED", "SEPARATE_VPC"}
)
