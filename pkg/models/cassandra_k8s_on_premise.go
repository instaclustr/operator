package models

type CassandraOnPremisesConfig struct {
	DataVolumeImageURL string
}

const (
	ApacheCassandra          = "APACHE_CASSANDRA"
	SSHDVPrefix              = "ssh-data-volume"
	SSHVMPrefix              = "ssh-virt-machine"
	SSHSvcPrefix             = "ssh-service"
	NodeSysDVPrefix          = "node-sys-data-volume"
	NodeDVPrefix             = "node-data-volume"
	NodeVMPrefix             = "node-virt-machine"
	NodeSvcPrefix            = "node-service"
	NodeIgnitionSecretFormat = "node-%d"
	IgnitionSecretData       = "script"
	LBType                   = "LoadBalancer"
	SSH                      = "ssh"
	DVDisk                   = "dvdisk"
	DataDisk                 = "datadisk"
	IgnitionDisk             = "ignition"
	DataDiskSerial           = "DATADISK"
	IgnitionSerial           = "IGNITION"
	SATA                     = "sata"
	Default                  = "default"
	CPU                      = "cpu"
	Memory                   = "memory"
	CloudInitDisk            = "cloudinitdisk"
	CloudInitSecretName      = "instaclustr-cloud-init-secret"
	Storage                  = "storage"
	DVKind                   = "DataVolume"
	OnPremisesDC             = "CLIENT_DC"
	OnPremisesNetwork        = "10.1.0.0/16"
	OnPremisesNodeSize       = "CAS-PRD-OP.2.2-200"
	OnPremisesProvider       = "ONPREMISES"
	IPModifyNeeded           = "IPModifyNeeded"
	MsgIPModifyNeeded        = "Please contact an Instaclustr support to apply cluster nodes IPs in the Instaclustr system"
	IgnitionScriptNeeded     = "ignitionScriptNeeded"
	MsgIgnitionScriptNeeded  = "Please contact an Instaclustr support to get ignition scripts for cluster and add them to secrets, then put secrets names to ignitionScriptsSecretNames in on-premises spec. Format: 'ssh: <ssh_secret>', 'node-1: <first_node_secret>'. Data format in secret need to be 'script: <base64_ecnoded_ignition_script>'"
)

type CassandraClusterV1 struct {
	ID                    string               `json:"id,omitempty"`
	ClusterStatus         string               `json:"clusterStatus,omitempty"`
	ClusterName           string               `json:"clusterName"`
	Bundles               []*CassandraBundle   `json:"bundles,omitempty"`
	Provider              *ClusterProviderV1   `json:"provider,omitempty"`
	PrivateNetworkCluster bool                 `json:"privateNetworkCluster,omitempty"`
	PCICompliantCluster   bool                 `json:"pciCompliantCluster,omitempty"`
	PrivateLink           []*PrivateLink       `json:"privateLink,omitempty"`
	SLATier               string               `json:"slaTier,omitempty"`
	NodeSize              string               `json:"nodeSize,omitempty"`
	ClusterNetwork        string               `json:"clusterNetwork,omitempty"`
	DataCentre            string               `json:"dataCentre,omitempty"`
	DataCentreCustomName  string               `json:"dataCentreCustomName,omitempty"`
	DataCentres           []*DataCentreV1      `json:"dataCentres,omitempty"`
	RackAllocation        *RackAllocationV1    `json:"rackAllocation,omitempty"`
	FirewallRules         []*FirewallRule      `json:"firewallRules,omitempty"`
	TwoFactorDelete       []*TwoFactorDeleteV1 `json:"twoFactorDelete,omitempty"`
	OIDCProvider          string               `json:"oidcProvider,omitempty"`
	BundledUseOnlyCluster bool                 `json:"bundledUseOnlyCluster,omitempty"`
	StorageNetwork        string               `json:"storageNetwork,omitempty"`
}

type CassandraBundleOptions struct {
	AuthnAuthz                    bool `json:"authnAuthz"`
	ClientEncryption              bool `json:"clientEncryption"`
	UsePrivateBroadcastRPCAddress bool `json:"usePrivateBroadcastRPCAddress"`
	LuceneEnabled                 bool `json:"luceneEnabled"`
	ContinuousBackupEnabled       bool `json:"continuousBackupEnabled"`
}

type CassandraBundle struct {
	Bundle  string                  `json:"bundle"`
	Version string                  `json:"version"`
	Options *CassandraBundleOptions `json:"options,omitempty"`
}

var (
	Port22 int32 = 22
)
