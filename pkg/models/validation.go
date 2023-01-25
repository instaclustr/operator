package models

var (
	CassandraVersions  = []string{"4.0.4", "3.11.13"}
	SparkVersions      = []string{"2.3.2", "3.0.1"}
	PostgreSQLVersions = []string{"15.1.0", "14.6.0", "14.5.0", "13.9.0", "13.8.0"}
	PGBouncerVersions  = []string{"1.17.0"}
	PoolModes          = []string{"TRANSACTION", "SESSION", "STATEMENT"}
	ReplicationModes   = []string{"ASYNCHRONOUS", "SYNCHRONOUS"}
	CloudProviders     = []string{"AWS_VPC", "GCP", "AZURE", "AZURE_AZ"}
	SLATiers           = []string{"PRODUCTION", "NON_PRODUCTION"}
	ClusterNameRegExp  = "^[a-zA-Z0-9][a-zA-Z0-9_-]{2,31}$"
)
