package models

var (
	RedisVersions      = []string{"6.2.7", "7.0.5"}
	CassandraVersions  = []string{"4.0.4", "3.11.13"}
	SparkVersions      = []string{"2.3.2", "3.0.1"}
	PostgreSQLVersions = []string{"15.1.0", "14.6.0", "14.5.0", "13.9.0", "13.8.0"}
	PGBouncerVersions  = []string{"1.17.0"}
	OpenSearchVersions = []string{"opensearch:1.3.7", "opensearch:2.2.1", "opensearch:1.3.7.ic1", "opensearch:2.2.1.ic1"}
	PoolModes          = []string{"TRANSACTION", "SESSION", "STATEMENT"}
	ReplicationModes   = []string{"ASYNCHRONOUS", "SYNCHRONOUS"}
	CloudProviders     = []string{"AWS_VPC", "GCP", "AZURE", "AZURE_AZ"}
	SLATiers           = []string{"PRODUCTION", "NON_PRODUCTION"}
	ClusterNameRegExp  = "^[a-zA-Z0-9][a-zA-Z0-9_-]{2,31}$"
)
