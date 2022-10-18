package models

import (
	"encoding/json"
)

const (
	AWSVPC  = "AWS_VPC"
	GCP     = "GCP"
	AZURE   = "AZURE"
	AZUREAZ = "AZURE_AZ"
)

type PatchRequest struct {
	Operation string          `json:"op"`
	Path      string          `json:"path"`
	Value     json.RawMessage `json:"value"`
}

type Cluster struct {
	Name                  string             `json:"name"`
	SLATier               string             `json:"slaTier"`
	PrivateNetworkCluster bool               `json:"privateNetworkCluster"`
	PCIComplianceMode     bool               `json:"pciComplianceMode"`
	TwoFactorDeletes      []*TwoFactorDelete `json:"twoFactorDelete,omitempty"`
}

type DataCentre struct {
	Name                string          `json:"name"`
	Network             string          `json:"network"`
	NodeSize            string          `json:"nodeSize"`
	NumberOfNodes       int32           `json:"numberOfNodes"`
	AWSSettings         []*AWSSetting   `json:"awsSettings,omitempty"`
	GCPSettings         []*GCPSetting   `json:"gcpSettings,omitempty"`
	AzureSettings       []*AzureSetting `json:"azureSettings,omitempty"`
	Tags                []*Tag          `json:"tags,omitempty"`
	CloudProvider       string          `json:"cloudProvider"`
	Region              string          `json:"region"`
	ProviderAccountName string          `json:"providerAccountName,omitempty"`
}

type AWSSetting struct {
	EBSEncryptionKey       string `json:"ebsEncryptionKey,omitempty"`
	CustomVirtualNetworkID string `json:"customVirtualNetworkId,omitempty"`
}

type GCPSetting struct {
	CustomVirtualNetworkID string `json:"customVirtualNetworkid,omitempty"`
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
	Rack           string   `json:"rack,omitempty"`
}

type ClusterStatus struct {
	ID          string              `json:"id,omitempty"`
	Status      string              `json:"status,omitempty"`
	DataCentres []*DataCentreStatus `json:"dataCentres,omitempty"`
}

type DataCentreStatus struct {
	ID            string  `json:"id"`
	Status        string  `json:"status"`
	Nodes         []*Node `json:"nodes"`
	NumberOfNodes int32   `json:"numberOfNodes"`
}
