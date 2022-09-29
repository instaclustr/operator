package instaclustr

type ClusterSpecAPIv1 struct {
	ClusterName           string          `json:"clusterName"`
	Provider              *ClusterProviderAPIv1 `json:"provider"`
	PrivateNetworkCluster bool            `json:"privateNetworkCluster,omitempty"`
	SLATier               string          `json:"slaTier,omitempty"`
	NodeSize              string          `json:"nodeSize"`
	ClusterNetwork        string          `json:"clusterNetwork,omitempty"`

	// DataCentre is a single data centre, for multiple leave blank and use DataCentres.
	DataCentre            string           `json:"dataCentre,omitempty"`
	DataCentreCustomName  string           `json:"dataCentreCustomName,omitempty"`
	RackAllocation        *RackAllocationAPIv1  `json:"rackAllocation,omitempty"`
	FirewallRules         []*FirewallRuleAPIv1  `json:"firewallRules,omitempty"`
	TwoFactorDelete       *TwoFactorDeleteAPIv1 `json:"twoFactorDelete,omitempty"`
	OIDCProvider          string           `json:"oidcProvider,omitempty"`
	BundledUseOnlyCluster bool             `json:"bundledUseOnlyCluster,omitempty"`
}

type ClusterProviderAPIv1 struct {
	Name                   string            `json:"name"`
	AccountName            string            `json:"accountName,omitempty"`
	CustomVirtualNetworkId string            `json:"customVirtualNetworkId,omitempty"`
	Tags                   map[string]string `json:"tags,omitempty"`
	ResourceGroup          string            `json:"resourceGroup,omitempty"`
	DiskEncryptionKey      string            `json:"diskEncryptionKey,omitempty"`
}

type RackAllocationAPIv1 struct {
	NumberOfRacks int32 `json:"numberOfRacks"`
	NodesPerRack  int32 `json:"nodesPerRack"`
}

type FirewallRuleAPIv1 struct {
	Network         string     `json:"network,omitempty"`
	SecurityGroupId string     `json:"securityGroupId,omitempty"`
	Rules           []*RuleTypeAPIv1 `json:"rules"`
}

type RuleTypeAPIv1 struct {
	Type string `json:"type"`
}

type TwoFactorDeleteAPIv1 struct {
	DeleteVerifyEmail string `json:"deleteVerifyEmail,omitempty"`
	DeleteVerifyPhone string `json:"deleteVerifyPhone,omitempty"`
}

type BundleAPIv1 struct {
	Bundle  string `json:"bundle"`
	Version string `json:"version"`
}

type DataCentreAPIv1 struct {
	Name           string           `json:"name,omitempty"`
	DataCentre     string           `json:"dataCentre"`
	Network        string           `json:"network"`
	Provider       *ClusterProviderAPIv1 `json:"provider,omitempty"`
	NodeSize       string           `json:"nodeSize,omitempty"`
	RackAllocation *RackAllocationAPIv1  `json:"rackAllocation,omitempty"`
}

type ClusterStatusAPIv1 struct {
	ID                  string `json:"id,omitempty"`
	ClusterCertificateDownload string `json:"clusterCertificateDownload,omitempty"`

	// ClusterStatus shows cluster current state such as a RUNNING, PROVISIONED, FAILED, etc.
	ClusterStatus string `json:"clusterStatus,omitempty"`
}

type PostgreSQLBundleAPIv1 struct {
	BundleAPIv1  `json:",inline"`
	Options *PostgreSQLBundleOptionsAPIv1 `json:"options"`
}

type PostgreSQLBundleOptionsAPIv1 struct {
	// PostgreSQL
	ClientEncryption      bool   `json:"clientEncryption,omitempty"`
	PostgresqlNodeCount   int32  `json:"postgresqlNodeCount"`
	ReplicationMode       string `json:"replicationMode,omitempty"`
	SynchronousModeStrict bool   `json:"synchronousModeStrict,omitempty"`
	// PGBouncer
	PoolMode string `json:"poolMode,omitempty"`
}

type PostgreSQLDataCentreAPIv1 struct {
	DataCentreAPIv1 `json:",inline"`
	Bundles    []*PostgreSQLBundleAPIv1 `json:"bundles"`
}

type PostgreSQLClusterAPIv1 struct {
	ClusterSpecAPIv1 `json:",inline"`
	Bundles     []*PostgreSQLBundleAPIv1     `json:"bundles"`
	DataCentres []*PostgreSQLDataCentreAPIv1 `json:"dataCentres,omitempty"`
}

type PostgreSQLStatusAPIv1 struct {
	ClusterStatusAPIv1 `json:",inline"`
	DataCentres []*PostgreSQLDataCentreAPIv1 `json:"dataCentres,omitempty"`
}
