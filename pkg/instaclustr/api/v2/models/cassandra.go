package models

// AWS nodes
const (
	CAS_DEV_t4g_small_5        = "CAS-DEV-t4g.small-5"
	CAS_DEV_t4g_small_30       = "CAS-DEV-t4g.small-30"

	CAS_DEV_t4g_medium_30      = "CAS-DEV-t4g.medium-30"
	CAS_DEV_t4g_medium_80      = "CAS-DEV-t4g.medium-80"

	CAS_PRD_m6g_large_120      = "CAS-PRD-m6g.large-120"
	CAS_PRD_m6g_large_250      = "CAS-PRD-m6g.large-250"
	CAS_PRD_m6g_large_400      = "CAS-PRD-m6g.large-400"
	CAS_PRD_m6g_large_600      = "CAS-PRD-m6g.large-600"

	CAS_PRD_r6g_large_250      = "CAS-PRD-r6g.large-250"
	CAS_PRD_r6g_large_400      = "CAS-PRD-r6g.large-400"
	CAS_PRD_r6g_large_600      = "CAS-PRD-r6g.large-600"
	CAS_PRD_r6g_large_800      = "CAS-PRD-r6g.large-800"
	CAS_PRD_r6g_large_1200     = "CAS-PRD-r6g.large-1200"
	CAS_PRD_r6g_large_1600     = "CAS-PRD-r6g.large-1600"

	CAS_PRD_r6g_xlarge_400     = "CAS-PRD-r6g.xlarge-400"
	CAS_PRD_r6g_xlarge_600     = "CAS-PRD-r6g.xlarge-600"
	CAS_PRD_r6g_xlarge_800     = "CAS-PRD-r6g.xlarge-800"
	CAS_PRD_r6g_xlarge_1200    = "CAS-PRD-r6g.xlarge-1200"
	CAS_PRD_r6g_xlarge_1600    = "CAS-PRD-r6g.xlarge-1600"
	CAS_PRD_r6g_xlarge_2400    = "CAS-PRD-r6g.xlarge-2400"
	CAS_PRD_r6g_xlarge_3200    = "CAS-PRD-r6g.xlarge-3200"

	CAS_PRD_r6g_2xlarge_800    = "CAS-PRD-r6g.2xlarge-800"
	CAS_PRD_r6g_2xlarge_1200   = "CAS-PRD-r6g.2xlarge-1200"
	CAS_PRD_r6g_2xlarge_1600   = "CAS-PRD-r6g.2xlarge-1600"
	CAS_PRD_r6g_2xlarge_2400   = "CAS-PRD-r6g.2xlarge-2400"
	CAS_PRD_r6g_2xlarge_3200   = "CAS-PRD-r6g.2xlarge-3200"

	CAS_PRD_r6g_4xlarge_800    = "CAS-PRD-r6g.4xlarge-800"
	CAS_PRD_r6g_4xlarge_1200   = "CAS-PRD-r6g.4xlarge-1200"
	CAS_PRD_r6g_4xlarge_1600   = "CAS-PRD-r6g.4xlarge-1600"
	CAS_PRD_r6g_4xlarge_2400   = "CAS-PRD-r6g.4xlarge-2400"
	CAS_PRD_r6g_4xlarge_3200   = "CAS-PRD-r6g.4xlarge-3200"

	CAS_PRD_i3_4xlarge_3538    = "CAS-PRD-i3.4xlarge-3538"
	CAS_PRD_i3en_2xlarge_4656  = "CAS-PRD-i3en.2xlarge-4656"
	c5d_2xlarge_v2             = "c5d.2xlarge_v2"
	i3_2xlarge_v2              = "i3.2xlarge-v2"
	i3en_xlarge                = "i3en.xlarge"
	c5d_2xlarge                = "c5d.2xlarge"
	i3_2xlarge                 = "i3.2xlarge"
	
)

// Azure nodes
const (
	D15_v2_an                = "D15_v2-an"
	L8s_v2_an                = "L8s_v2-an"
	Standard_DS12_v2_1023_an = "Standard_DS12_v2-1023-an"
	Standard_DS12_v2_2046_an = "Standard_DS12_v2-2046-an"
	Standard_DS12_v2_512_an  = "Standard_DS12_v2-512-an"
	Standard_DS13_v2_2046_an = "Standard_DS13_v2-2046-an"
	Standard_DS2_v2_256_an   = "Standard_DS2_v2-256-an"
	Standard_DS12_v2_1023    = "Standard_DS12_v2-1023"
	Standard_DS12_v2_2046    = "Standard_DS12_v2-2046"
	Standard_DS12_v2_512     = "Standard_DS12_v2-512"
	Standard_DS13_v2_2046    = "Standard_DS13_v2-2046"
	Standard_DS2_v2_256      = "Standard_DS2_v2-256"
)

// GCP nodes
const (
	cassandra_production_n2_highmem_16_3000 = "cassandra-production-n2-highmem-16-3000"
	cassandra_production_n2_highmem_4_3000  = "cassandra-production-n2-highmem-4-3000"
	cassandra_production_n2_highmem_8_1500  = "cassandra-production-n2-highmem-8-1500"
	cassandra_production_n2_standard_8_375  = "cassandra-production-n2-standard-8-375"
	cassandra_production_n2_standard_8_750  = "cassandra-production-n2-standard-8-750"
	n1_highmem_4_1600                       = "n1-highmem-4-1600"
	n1_highmem_4_400                        = "n1-highmem-4-400"
	n1_highmem_4_800                        = "n1-highmem-4-800"
	n1_standard_1                           = "n1-standard-1"
	n1_standard_2                           = "n1-standard-2"
	n1_standard_4_1600                      = "n1-standard-4-1600"
	n1_standard_4_400                       = "n1-standard-4-400"
	n1_standard_4_800                       = "n1-standard-4-800"
)

var awsCassandraNodes = map[string]int{
	CAS_DEV_t4g_small_5
	CAS_DEV_t4g_small_30
	CAS_DEV_t4g_medium_30
	CAS_DEV_t4g_medium_80
	CAS_PRD_m6g_large_120
	CAS_PRD_m6g_large_250
	CAS_PRD_m6g_large_400
	CAS_PRD_m6g_large_600
	CAS_PRD_r6g_large_250
	CAS_PRD_r6g_large_400
	CAS_PRD_r6g_large_600
	CAS_PRD_r6g_large_800
	CAS_PRD_r6g_large_1200
	CAS_PRD_r6g_large_1600
	CAS_PRD_r6g_xlarge_400
	CAS_PRD_r6g_xlarge_600
	CAS_PRD_r6g_xlarge_800
	CAS_PRD_r6g_xlarge_1200
	CAS_PRD_r6g_xlarge_1600
	CAS_PRD_r6g_xlarge_2400
	CAS_PRD_r6g_xlarge_3200
	CAS_PRD_r6g_2xlarge_800
	CAS_PRD_r6g_2xlarge_1200
	CAS_PRD_r6g_2xlarge_1600
	CAS_PRD_r6g_2xlarge_2400
	CAS_PRD_r6g_2xlarge_3200
	CAS_PRD_r6g_4xlarge_800
	CAS_PRD_r6g_4xlarge_1200
	CAS_PRD_r6g_4xlarge_1600
	CAS_PRD_r6g_4xlarge_2400
	CAS_PRD_r6g_4xlarge_3200
}

type CassandraCluster struct {
	Cluster             `json:",inline"`
	DataCentres         []*CassandraDataCentre `json:"dataCentres"`
	CassandraVersion    string                 `json:"cassandraVersion"`
	LuceneEnabled       bool                   `json:"luceneEnabled"`
	PasswordAndUserAuth bool                   `json:"passwordAndUserAuth"`
	Spark               []*Spark               `json:"spark,omitempty"`
}

type CassandraDataCentre struct {
	DataCentre                     `json:",inline"`
	ReplicationFactor              int32 `json:"replicationFactor"`
	ContinuousBackup               bool  `json:"continuousBackup"`
	PrivateIPBroadcastForDiscovery bool  `json:"privateIpBroadcastForDiscovery"`
	ClientToClusterEncryption      bool  `json:"clientToClusterEncryption"`
}

// CassandraStatus defines the observed state of Cassandra
type CassandraStatus struct {
	ClusterStatus   `json:",inline"`
	OperationStatus string `json:"operationStatus,omitempty"`
}

type CassandraDCs struct {
	DataCentres []*CassandraDataCentre `json:"dataCentres"`
}

type Spark struct {
	Version string `json:"version"`
}
