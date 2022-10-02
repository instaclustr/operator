package models

const (
	PgSQL     = "POSTGRESQL"
	PgBouncer = "PGBOUNCER"
)

type PgBundleAPIv1 struct {
	BundleAPIv1 `json:",inline"`
	Options     *PgBundleOptionsAPIv1 `json:"options"`
}

type PgBundleOptionsAPIv1 struct {
	// PostgreSQL
	ClientEncryption      bool   `json:"clientEncryption,omitempty"`
	PostgresqlNodeCount   int32  `json:"postgresqlNodeCount"`
	ReplicationMode       string `json:"replicationMode,omitempty"`
	SynchronousModeStrict bool   `json:"synchronousModeStrict,omitempty"`
	// PGBouncer
	PoolMode string `json:"poolMode,omitempty"`
}

type PgDataCentreAPIv1 struct {
	DataCentreAPIv1 `json:",inline"`
	Bundles         []*PgBundleAPIv1 `json:"bundles"`
}

type PgClusterAPIv1 struct {
	ClusterAPIv1 `json:",inline"`
	Bundles      []*PgBundleAPIv1     `json:"bundles"`
	DataCentres  []*PgDataCentreAPIv1 `json:"dataCentres,omitempty"`
}

type PgStatusAPIv1 struct {
	ClusterStatusAPIv1 `json:",inline"`
	DataCentres        []*DataCentreStatusAPIv1 `json:"dataCentres,omitempty"`
}
