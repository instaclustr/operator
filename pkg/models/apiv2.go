package models

const (
	NoOperation = "NO_OPERATION"

	DefaultAccountName = "INSTACLUSTR"

	AWSVPC  = "AWS_VPC"
	GCP     = "GCP"
	AZURE   = "AZURE"
	AZUREAZ = "AZURE_AZ"
)

type ClusterStatus struct {
	ID                            string `json:"id,omitempty"`
	Status                        string `json:"status,omitempty"`
	CurrentClusterOperationStatus string `json:"currentClusterOperationStatus,omitempty"`
}

type DataCentre struct {
	DataCentreStatus    `json:",inline"`
	Name                string          `json:"name"`
	Network             string          `json:"network"`
	NodeSize            string          `json:"nodeSize"`
	NumberOfNodes       int             `json:"numberOfNodes,omitempty"`
	AWSSettings         []*AWSSetting   `json:"awsSettings,omitempty"`
	GCPSettings         []*GCPSetting   `json:"gcpSettings,omitempty"`
	AzureSettings       []*AzureSetting `json:"azureSettings,omitempty"`
	Tags                []*Tag          `json:"tags,omitempty"`
	CloudProvider       string          `json:"cloudProvider"`
	Region              string          `json:"region"`
	ProviderAccountName string          `json:"providerAccountName,omitempty"`
}

type DataCentreStatus struct {
	ID     string  `json:"id,omitempty"`
	Status string  `json:"status,omitempty"`
	Nodes  []*Node `json:"nodes,omitempty"`
}

type CloudProviderSettings struct {
	AWSSettings   []*AWSSetting   `json:"awsSettings,omitempty"`
	GCPSettings   []*GCPSetting   `json:"gcpSettings,omitempty"`
	AzureSettings []*AzureSetting `json:"azureSettings,omitempty"`
}

type AWSSetting struct {
	EBSEncryptionKey       string `json:"ebsEncryptionKey,omitempty"`
	CustomVirtualNetworkID string `json:"customVirtualNetworkId,omitempty"`
}

type GCPSetting struct {
	CustomVirtualNetworkID string `json:"customVirtualNetworkId,omitempty"`
}

type AzureSetting struct {
	ResourceGroup string `json:"resourceGroup,omitempty"`
}

type Tag struct {
	Value string `json:"value"`
	Key   string `json:"key"`
}

type Node struct {
	ID             string   `json:"id,omitempty"`
	Size           string   `json:"nodeSize,omitempty"`
	Status         string   `json:"status,omitempty"`
	Roles          []string `json:"nodeRoles,omitempty"`
	PublicAddress  string   `json:"publicAddress,omitempty"`
	PrivateAddress string   `json:"privateAddress,omitempty"`
	Rack           string   `json:"rack,omitempty"`
}

type TwoFactorDelete struct {
	ConfirmationPhoneNumber string `json:"confirmationPhoneNumber,omitempty"`
	ConfirmationEmail       string `json:"confirmationEmail"`
}

type NodeReloadStatus struct {
	NodeID       string `json:"nodeId,omitempty"`
	OperationID  string `json:"operationId,omitempty"`
	TimeCreated  string `json:"timeCreated,omitempty"`
	TimeModified string `json:"timeModified,omitempty"`
	Status       string `json:"status,omitempty"`
	Message      string `json:"message,omitempty"`
}

type ActiveClusters struct {
	AccountID string           `json:"accountId,omitempty"`
	Clusters  []*ActiveCluster `json:"clusters,omitempty"`
}

type ActiveCluster struct {
	Application string `json:"application,omitempty"`
	ID          string `json:"id,omitempty"`
}

type AppVersions struct {
	Application string   `json:"application"`
	Versions    []string `json:"versions"`
}
