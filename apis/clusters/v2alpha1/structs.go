package v2alpha1

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

type DataCentre struct {
	// A logical Name for the data centre within a cluster. These names must be unique in the cluster.
	Name string `json:"name"`

	// The private network address block for the Data Centre specified using CIDR address notation.
	// The Network must have a prefix length between /12 and /22 and must be part of a private address space.
	Network string `json:"network"`

	// NodeSize is a size of the nodes provisioned in the Data Centre.
	NodeSize string `json:"nodeSize"`

	// Total number of Kafka brokers in the Data Centre. Must be a multiple of defaultReplicationFactor.
	NumberOfNodes int32 `json:"numberOfNodes"`

	// AWS specific settings for the Data Centre. Cannot be provided with GCPSettings and AzureSettings.
	AWSSettings []*AWSSettings `json:"awsSettings,omitempty"`

	// GCPSettings specific settings for the Data Centre. Cannot be provided with AWSSettings and AzureSettings.
	GCPSettings []*GCPSettings `json:"gcpSettings,omitempty"`

	// AzureSettings specific settings for the Data Centre. Cannot be provided with GCPSettings and AWSSettings.
	AzureSettings []*AzureSettings `json:"azureSettings,omitempty"`

	// List of tags to apply to the Data Centre. Tags are metadata labels which allow you to identify,
	// categorize and filter clusters. This can be useful for grouping together clusters into applications,
	// environments, or any category that you require.
	Tags []*Tag `json:"tags,omitempty"`

	// Enum: "AWS_VPC" "GCP" "AZURE" "AZURE_AZ"
	// CloudProvider is name of the cloud provider service in which the Data Centre will be provisioned.
	CloudProvider string `json:"cloudProvider"`

	// Region of the Data Centre.
	Region string `json:"region"`

	// For customers running in their own account. Your provider account can be found on the Create Cluster page on
	// the Instaclustr Console, or the "Provider Account" property on any existing cluster. For customers provisioning on
	// Instaclustr's cloud provider accounts, this property may be omitted.
	ProviderAccountName string `json:"providerAccountName,omitempty"`
}

type Node struct {
	ID             string   `json:"id,omitempty"`
	Size           string   `json:"nodeSize,omitempty"`
	Status         string   `json:"status,omitempty"`
	Roles          []string `json:"nodeRoles,omitempty"`
	Rack           string   `json:"rack,omitempty"`
	PublicAddress  string   `json:"publicAddress,omitempty"`
	PrivateAddress string   `json:"privateAddress,omitempty"`
}

type TwoFactorDelete struct {
	ConfirmationPhoneNumber string `json:"confirmationPhoneNumber,omitempty"`
	ConfirmationEmail       string `json:"confirmationEmail"`
}

type Tag struct {

	// Value of the tag to be added to the Data Centre.
	Value string `json:"value"`

	// Key of the tag to be added to the Data Centre.
	Key string `json:"key"`
}

type ClusterSpec struct {
	Name                  string             `json:"name"`
	SLATier               string             `json:"slaTier"`
	PrivateNetworkCluster bool               `json:"privateNetworkCluster"`
	PCIComplianceMode     bool               `json:"pciComplianceMode"`
	TwoFactorDelete       []*TwoFactorDelete `json:"twoFactorDelete,omitempty"`
}
