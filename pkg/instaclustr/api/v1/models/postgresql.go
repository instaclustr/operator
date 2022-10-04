package models

const (
	PgSQL     = "POSTGRESQL"
	PgBouncer = "PGBOUNCER"
)

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
