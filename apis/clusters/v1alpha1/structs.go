package v1alpha1

type DataCentre struct {
	Name           string           `json:"name,omitempty"`
	Region     string           `json:"Region"`
	Network        string           `json:"network"`
	CloudProvider       string `json:"cloudProvider,omitempty"`
	NumberOfNodes int32 `json:"numberOfNodes"`
	NodeSize string `json:"nodeSize"`
	ProviderAccountName string `json:"providerAccountName,omitempty"`
	AWSSettings []*AWSSettings `json:"awsSettings,omitempty"`
	GCPSettings []*GCPSettings `json:"gcpSettings,omitempty"`
	AzureSettings []*AzureSettings `json:"azureSettings,omitempty"`
	Tags []*Tag `json:"tags,omitempty"`
}

type AWSSettings struct {
	// ID of a KMS encryption key to encrypt data on nodes.
	// KMS encryption key must be set in Cluster Resources through the Instaclustr Console before
	// provisioning an encrypted Data Centre.
	EBSEncryptionKey string `json:"ebsEncryptionKey,omitempty"`

	// VPC ID into which the Data Centre will be provisioned.
	// The Data Centre's network allocation must match the IPv4 CIDR block of the specified VPC.
	CustomVirtualNetworkID string `json:"customVirtualNetworkId,omitempty"`
}

type GCPSettings struct {
	// Network name or a relative Network or Subnetwork URI e.g. projects/my-project/regions/us-central1/subnetworks/my-subnet.
	// The Data Centre's network allocation must match the IPv4 CIDR block of the specified subnet.
	CustomVirtualNetworkID string `json:"customVirtualNetworkId,omitempty"`
}

type AzureSettings struct {
	ResourceGroup string `json:"resourceGroup,omitempty"`
}

type Tag struct {
	// Value of the tag to be added to the Data Centre.
	Value string `json:"value"`

	// Key of the tag to be added to the Data Centre.
	Key string `json:"key"`
}

type DataCentreStatus struct {
	ID        string  `json:"id"`
	Status    string  `json:"status"`
	Nodes     []*Node `json:"nodes"`
}

type Node struct {
	ID             string `json:"id"`
	Size       string           `json:"nodeSize,omitempty"`
	Status         string `json:"status"`
	Rack           string `json:"rack"`
	PublicAddress  string `json:"publicAddress"`
	PrivateAddress string `json:"privateAddress"`
}

type ClusterSpec struct {
	Name           string          `json:"Name"`
	SLATier               string          `json:"slaTier,omitempty"`
	PrivateNetworkCluster bool            `json:"privateNetworkCluster,omitempty"`
	TwoFactorDelete       *[]TwoFactorDelete `json:"twoFactorDelete,omitempty"`
	OIDCProvider          string           `json:"oidcProvider,omitempty"`
	BundledUseOnlyCluster bool             `json:"bundledUseOnlyCluster,omitempty"`
}

type ClusterStatus struct {
	ID                  string `json:"id,omitempty"`
	ClusterCertificateDownload string `json:"clusterCertificateDownload,omitempty"`

	// Status shows cluster current state such as a RUNNING, PROVISIONED, FAILED, etc.
	Status string `json:"status,omitempty"`
}

type FirewallRule struct {
	Network         string     `json:"network,omitempty"`
	SecurityGroupId string     `json:"securityGroupId,omitempty"`
	Rules           []RuleType `json:"rules"`
}

type RuleType struct {
	Type string `json:"type"`
}

type TwoFactorDelete struct {
	ConfirmationPhoneNumber string `json:"confirmationPhoneNumber,omitempty"`
	ConfirmationEmail string `json:"confirmationEmail,omitempty"`
}
