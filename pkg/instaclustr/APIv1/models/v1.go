package models

const (
	ClustersCreationEndpoint = "/provisioning/v1/extended/"

	// ClustersEndpoint is used for GET, DELETE and UPDATE clusters
	ClustersEndpoint = "/provisioning/v1/"

	// ClustersResizeEndpoint is used for nodes resizing in a cluster data centre.
	// Example: fmt.Sprintf("%s/provisioning/v1/%s/%s/resize", serverHostname, clusterID, dataCentreID)
	ClustersResizeEndpoint = "%s/provisioning/v1/%s/%s/resize"
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
	ID                         string `json:"id,omitempty"`
	ClusterCertificateDownload string `json:"clusterCertificateDownload,omitempty"`

	// ClusterStatus shows cluster current state such as a RUNNING, PROVISIONED, FAILED, etc.
	ClusterStatus string `json:"clusterStatus,omitempty"`
}

type DataCentreStatus struct {
	ID        string        `json:"id"`
	Status    string        `json:"status"`
	Nodes     []*NodeStatus `json:"nodes"`
	NodeCount int32         `json:"nodeCount"`
}

type NodeStatus struct {
	ID             string `json:"id"`
	Rack           string `json:"rack"`
	Size           string `json:"size"`
	PublicAddress  string `json:"publicAddress"`
	PrivateAddress string `json:"privateAddress"`
	Status         string `json:"status"`
}
