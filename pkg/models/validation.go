package models

var (
	CassandraVersions = []string{"4.0.4", "3.11.13"}
	SparkVersions     = []string{"2.3.2", "3.0.1"}
	CloudProviders    = []string{"AWS_VPC", "GCP", "AZURE", "AZURE_AZ"}
	SLATiers          = []string{"PRODUCTION", "NON_PRODUCTION"}
	ClusterNameRegExp = "^[a-zA-Z0-9][a-zA-Z0-9_-]{2,31}$"
)
