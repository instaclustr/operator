package models

const (
	PgSQL         = "POSTGRESQL"
	PgBouncer     = "PGBOUNCER"
	PgNodePurpose = "POSTGRESQL"
)

const (
	// AWS node types
	pgDev_t4g_small_5      = "PGS-DEV-t4g.small-5"
	pgDev_t4g_medium_30    = "PGS-DEV-t4g.medium-30"
	pgPrd_m6g_large_250    = "PGS-PRD-m6g.large-250"
	pgPrd_m6g_large_500    = "PGS-PRD-m6g.large-500"
	pgPrd_m6g_large_1000   = "PGS-PRD-m6g.large-1000"
	pgPRD_m6g_xlarge_500   = "PGS-PRD-m6g.xlarge-500"
	pgPRD_m6g_xlarge_1000  = "PGS-PRD-m6g.xlarge-1000"
	pgPRD_m6g_xlarge_2000  = "PGS-PRD-m6g.xlarge-2000"
	pgPRD_m6g_xlarge_4000  = "PGS-PRD-m6g.xlarge-4000"
	pgPRD_m6g_2xlarge_1000 = "PGS-PRD-m6g.2xlarge-1000"
	pgPRD_m6g_2xlarge_2000 = "PGS-PRD-m6g.2xlarge-2000"
	pgPRD_m6g_2xlarge_4000 = "PGS-PRD-m6g.2xlarge-4000"
	pgPRD_m6g_2xlarge_8000 = "PGS-PRD-m6g.2xlarge-8000"
)

const (
	// Azure node types
	pgDev_std_ds1_v2_5_an     = "PGS-DEV-Standard_DS1_v2-5-an"
	pgDev_std_ds1_v2_30_an    = "PGS-DEV-Standard_DS1_v2-30-an"
	pgPRD_std_ds2_v2_250_an   = "PGS-PRD-Standard_DS2_v2-250-an"
	pgPRD_std_ds2_v2_500_an   = "PGS-PRD-Standard_DS2_v2-500-an"
	pgPRD_std_ds11_v2_250_an  = "PGS-PRD-Standard_DS11_v2-250-an"
	pgPRD_std_ds11_v2_500_an  = "PGS-PRD-Standard_DS11_v2-500-an"
	pgPRD_std_ds12_v2_500_an  = "PGS-PRD-Standard_DS12_v2-500-an"
	pgPRD_std_ds12_v2_1000_an = "PGS-PRD-Standard_DS12_v2-1000-an"
	pgPRD_std_ds12_v2_2000_an = "PGS-PRD-Standard_DS12_v2-2000-an"
	pgPRD_std_ds13_v2_1000_an = "PGS-PRD-Standard_DS13_v2-1000-an"
	pgPRD_std_ds13_v2_2000_an = "PGS-PRD-Standard_DS13_v2-2000-an"
	pgPRD_std_ds13_v2_4000_an = "PGS-PRD-Standard_DS13_v2-4000-an"
)

const (
	// GCP node types
	pgDev_n2_std_2_5    = "PGS-DEV-n2-standard-2-5"
	pgDev_n2_std_2_30   = "PGS-DEV-n2-standard-2-30"
	pgPRD_n2_std_2_250  = "PGS-PRD-n2-standard-2-250"
	pgPRD_n2_std_4_500  = "PGS-PRD-n2-standard-4-500"
	pgPRD_n2_std_4_1000 = "PGS-PRD-n2-standard-4-1000"
	pgPRD_n2_std_8_2000 = "PGS-PRD-n2-standard-8-2000"
	pgPRD_n2_std_8_4000 = "PGS-PRD-n2-standard-8-4000"
)

var PgAWSNodeTypes = map[string]int{
	pgDev_t4g_small_5:      1,
	pgDev_t4g_medium_30:    2,
	pgPrd_m6g_large_250:    3,
	pgPrd_m6g_large_500:    4,
	pgPrd_m6g_large_1000:   5,
	pgPRD_m6g_xlarge_500:   6,
	pgPRD_m6g_xlarge_1000:  7,
	pgPRD_m6g_xlarge_2000:  8,
	pgPRD_m6g_xlarge_4000:  9,
	pgPRD_m6g_2xlarge_1000: 10,
	pgPRD_m6g_2xlarge_2000: 11,
	pgPRD_m6g_2xlarge_4000: 12,
	pgPRD_m6g_2xlarge_8000: 13,
}
var PgAzureNodeTypes = map[string]int{
	pgDev_std_ds1_v2_5_an:     1,
	pgDev_std_ds1_v2_30_an:    2,
	pgPRD_std_ds2_v2_250_an:   3,
	pgPRD_std_ds2_v2_500_an:   4,
	pgPRD_std_ds11_v2_250_an:  5,
	pgPRD_std_ds11_v2_500_an:  6,
	pgPRD_std_ds12_v2_500_an:  7,
	pgPRD_std_ds12_v2_1000_an: 8,
	pgPRD_std_ds12_v2_2000_an: 9,
	pgPRD_std_ds13_v2_1000_an: 10,
	pgPRD_std_ds13_v2_2000_an: 11,
	pgPRD_std_ds13_v2_4000_an: 12,
}
var PgGCPNodeTypes = map[string]int{
	pgDev_n2_std_2_5:    1,
	pgDev_n2_std_2_30:   2,
	pgPRD_n2_std_2_250:  3,
	pgPRD_n2_std_4_500:  4,
	pgPRD_n2_std_4_1000: 5,
	pgPRD_n2_std_8_2000: 6,
	pgPRD_n2_std_8_4000: 7,
}

type PgBundle struct {
	Bundle  `json:",inline"`
	Options *PgBundleOptions `json:"options"`
}

type PgBundleOptions struct {
	// PostgreSQL
	ClientEncryption      bool   `json:"clientEncryption,omitempty"`
	PostgresqlNodeCount   int32  `json:"postgresqlNodeCount"`
	ReplicationMode       string `json:"replicationMode,omitempty"`
	SynchronousModeStrict bool   `json:"synchronousModeStrict,omitempty"`
	// PGBouncer
	PoolMode string `json:"poolMode,omitempty"`
}

type PgDataCentre struct {
	DataCentre `json:",inline"`
	Bundles    []*PgBundle `json:"bundles"`
}

type PgCluster struct {
	Cluster     `json:",inline"`
	Bundles     []*PgBundle     `json:"bundles"`
	DataCentres []*PgDataCentre `json:"dataCentres,omitempty"`
}

type PgStatus struct {
	ClusterStatus `json:",inline"`
	DataCentres   []*DataCentreStatus `json:"dataCentres,omitempty"`
}
