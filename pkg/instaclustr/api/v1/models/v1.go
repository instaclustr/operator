package models

const (
	ProvisionedStatus = "PROVISIONED"
	RunningStatus     = "RUNNING"
)

type Cluster struct {
	ClusterName           string           `json:"clusterName"`
	Provider              *ClusterProvider `json:"provider"`
	PrivateNetworkCluster bool             `json:"privateNetworkCluster,omitempty"`
	SLATier               string           `json:"slaTier,omitempty"`
	NodeSize              string           `json:"nodeSize"`
	ClusterNetwork        string           `json:"clusterNetwork,omitempty"`

	// DataCentre is a single data centre, for multiple leave blank and use DataCentres.
	DataCentre            string           `json:"dataCentre,omitempty"`
	DataCentreCustomName  string           `json:"dataCentreCustomName,omitempty"`
	RackAllocation        *RackAllocation  `json:"rackAllocation,omitempty"`
	FirewallRules         []*FirewallRule  `json:"firewallRules,omitempty"`
	TwoFactorDelete       *TwoFactorDelete `json:"twoFactorDelete,omitempty"`
	OIDCProvider          string           `json:"oidcProvider,omitempty"`
	BundledUseOnlyCluster bool             `json:"bundledUseOnlyCluster,omitempty"`
}

type ClusterProvider struct {
	Name                   string            `json:"name"`
	AccountName            string            `json:"accountName,omitempty"`
	CustomVirtualNetworkId string            `json:"customVirtualNetworkId,omitempty"`
	Tags                   map[string]string `json:"tags,omitempty"`
	ResourceGroup          string            `json:"resourceGroup,omitempty"`
	DiskEncryptionKey      string            `json:"diskEncryptionKey,omitempty"`
}

type RackAllocation struct {
	NumberOfRacks int32 `json:"numberOfRacks"`
	NodesPerRack  int32 `json:"nodesPerRack"`
}

type FirewallRule struct {
	Network         string      `json:"network,omitempty"`
	SecurityGroupId string      `json:"securityGroupId,omitempty"`
	Rules           []*RuleType `json:"rules"`
}

type RuleType struct {
	Type string `json:"type"`
}

type TwoFactorDelete struct {
	DeleteVerifyEmail string `json:"deleteVerifyEmail,omitempty"`
	DeleteVerifyPhone string `json:"deleteVerifyPhone,omitempty"`
}

type Bundle struct {
	Bundle  string `json:"bundle"`
	Version string `json:"version"`
}

type DataCentre struct {
	Name           string           `json:"name,omitempty"`
	DataCentre     string           `json:"dataCentre"`
	Network        string           `json:"network"`
	Provider       *ClusterProvider `json:"provider,omitempty"`
	NodeSize       string           `json:"nodeSize,omitempty"`
	RackAllocation *RackAllocation  `json:"rackAllocation,omitempty"`
}

type ClusterStatus struct {
	ID string `json:"id,omitempty"`

	// ClusterStatus shows cluster current state such as a RUNNING, PROVISIONED, FAILED, etc.
	ClusterStatus string              `json:"clusterStatus,omitempty"`
	DataCentres   []*DataCentreStatus `json:"dataCentres,omitempty"`
}

type DataCentreStatus struct {
	ID              string        `json:"id,omitempty"`
	CDCStatus       string        `json:"cdcStatus,omitempty"`
	Nodes           []*NodeStatus `json:"nodes,omitempty"`
	NodeCount       int32         `json:"nodeCount,omitempty"`
	EncryptionKeyID string        `json:"encryptionKeyId,omitempty"`
}

type NodeStatus struct {
	ID             string `json:"id"`
	Rack           string `json:"rack"`
	Size           string `json:"size"`
	PublicAddress  string `json:"publicAddress"`
	PrivateAddress string `json:"privateAddress"`
	NodeStatus     string `json:"nodeStatus"`
}

type FullClusterStatus struct {
	ClusterName     string `json:"clusterName,omitempty"`
	ID              string `json:"id,omitempty"`
	ClusterStatus   string `json:"clusterStatus,omitempty"`
	TwoFactorDelete bool   `json:"twoFactorDelete,omitempty"`
	CDCID           string `json:"cdcid,omitempty"`

	DataCentres []*DataCentreStatus `json:"dataCentres,omitempty"`
}

type ResizeRequest struct {
	NewNodeSize           string `json:"newNodeSize"`
	ConcurrentResizes     int    `json:"concurrentResizes"`
	NotifySupportContacts bool   `json:"notifySupportContacts"`
	NodePurpose           string `json:"nodePurpose"`
}

type DataCentreResizeOperations struct {
	CDCID  string `json:"cdc"`
	Status string `json:"completedStatus"`
}

type ClusterConfigurations struct {
	ParameterName  string `json:"parameterName"`
	ParameterValue string `json:"parameterValue"`
}

type ClusterModifyRequest struct {
	TwoFactorDelete *TwoFactorDelete `json:"twoFactorDelete,omitempty"`
	Description     string           `json:"description,omitempty"`
}
