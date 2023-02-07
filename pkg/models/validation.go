package models

var (
	RedisVersions      = []string{"6.2.7", "7.0.5"}
	CassandraVersions  = []string{"4.0.4", "3.11.13"}
	SparkVersions      = []string{"2.3.2", "3.0.1"}
	PostgreSQLVersions = []string{"15.1.0", "14.6.0", "14.5.0", "13.9.0", "13.8.0"}
	PGBouncerVersions  = []string{"1.17.0"}
	KafkaVersions      = []string{"3.0.2", "3.1.2", "2.8.2"}
	OpenSearchVersions = []string{"opensearch:1.3.7", "opensearch:2.2.1", "opensearch:1.3.7.ic1", "opensearch:2.2.1.ic1"}
	PoolModes          = []string{"TRANSACTION", "SESSION", "STATEMENT"}
	ReplicationModes   = []string{"ASYNCHRONOUS", "SYNCHRONOUS"}
	CloudProviders     = []string{"AWS_VPC", "GCP", "AZURE", "AZURE_AZ"}
	SLATiers           = []string{"PRODUCTION", "NON_PRODUCTION"}
	ClusterNameRegExp  = "^[a-zA-Z0-9][a-zA-Z0-9_-]{2,31}$"
	BundleTypes        = []string{"APACHE_ZOOKEEPER", "CADENCE", "CADENCE_GRPC",
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
	PeerSubnetsRegExp      = "^((10|172|192))\\.([^0]{1,2}|[^0]\\d|0{1}|1\\d\\d|2[0-4]\\d|25[0-5])\\.([^0]{1,2}|[^0]\\d|0{1}|1\\d\\d|2[0-4]\\d|25[0-5])\\.(0)(\\/1[6-9]|\\/2[0-8])$"
	DataCentreIDRegExp     = "^[0-9a-fA-F]{8}\\b-[0-9a-fA-F]{4}\\b-[0-9a-fA-F]{4}\\b-[0-9a-fA-F]{4}\\b-[0-9a-fA-F]{12}$"
	AWSRegions             = []string{"AF_SOUTH_1", "AP_EAST_1", "AP_NORTHEAST_1", "AP_NORTHEAST_2", "AP_SOUTHEAST_1",
		"AP_SOUTHEAST_2", "AP_SOUTH_1", "CA_CENTRAL_1", "CN_NORTHWEST_1", "CN_NORTH_1", "EU_CENTRAL_1", "EU_NORTH_1",
		"EU_SOUTH_1", "EU_WEST_1", "EU_WEST_2", "EU_WEST_3", "ME_SOUTH_1", "SA_EAST_1", "US_EAST_1", "US_EAST_2",
		"US_WEST_1", "US_WEST_2"}
	AzureRegions = []string{"AUSTRALIA_EAST", "CANADA_CENTRAL", "CENTRAL_US", "EAST_US", "EAST_US_2", "NORTH_EUROPE",
		"SOUTHEAST_ASIA", "SOUTH_CENTRAL_US", "WEST_EUROPE", "WEST_US_2"}
	GCPRegions = []string{"asia-east1", "asia-northeast1", "asia-south1", "asia-southeast1", "australia-southeast1",
		"europe-north1", "europe-west1", "europe-west2", "europe-west3", "europe-west4", "europe-west6",
		"northamerica-northeast1", "southamerica-east1", "us-central1", "us-east1", "us-east4", "us-west1", "us-west2"}
)
