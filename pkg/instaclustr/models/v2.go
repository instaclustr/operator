package models

type ClusterSpec struct {
	Name                  string             `json:"name"`
	SLATier               string             `json:"slaTier"`
	PrivateNetworkCluster bool               `json:"privateNetworkCluster"`
	PCIComplianceMode     bool               `json:"pciComplianceMode"`
	TwoFactorDeletes      []*TwoFactorDelete `json:"twoFactorDelete,omitempty"`
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
	AWSSettings []*AWSSetting `json:"awsSettings,omitempty"`

	// GCPSettings specific settings for the Data Centre. Cannot be provided with AWSSettings and AzureSettings.
	GCPSettings []*GCPSetting `json:"gcpSettings,omitempty"`

	// AzureSettings specific settings for the Data Centre. Cannot be provided with GCPSettings and AWSSettings.
	AzureSettings []*AzureSetting `json:"azureSettings,omitempty"`

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

type AWSSetting struct {
	EBSEncryptionKey       string `json:"ebsEncryptionKey,omitempty"`
	CustomVirtualNetworkID string `json:"customVirtualNetworkID,omitempty"`
}

type GCPSetting struct {
	CustomVirtualNetworkID string `json:"customVirtualNetworkID,omitempty"`
}

type AzureSetting struct {
	ResourceGroup string `json:"resourceGroup,omitempty"`
}

type Tag struct {
	Value string `json:"value"`
	Key   string `json:"key"`
}

type TwoFactorDelete struct {
	ConfirmationPhoneNumber string `json:"confirmationPhoneNumber"`
	ConfirmationEmail       string `json:"confirmationEmail"`
}

type Node struct {
	ID             string   `json:"id,omitempty"`
	Size           string   `json:"nodeSize,omitempty"`
	Status         string   `json:"status,omitempty"`
	Roles          []string `json:"nodeRoles,omitempty"`
	PublicAddress  string   `json:"publicAddress,omitempty"`
	PrivateAddress string   `json:"privateAddress,omitempty"`
	Rack           string   `json:"rack"`
}
