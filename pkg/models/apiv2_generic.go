package models

type GenericClusterFields struct {
	ID                            string `json:"id,omitempty"`
	Status                        string `json:"status,omitempty"`
	CurrentClusterOperationStatus string `json:"currentClusterOperationStatus,omitempty"`

	Name                  string             `json:"name"`
	Description           string             `json:"description,omitempty"`
	PCIComplianceMode     bool               `json:"pciComplianceMode"`
	PrivateNetworkCluster bool               `json:"privateNetworkCluster"`
	SLATier               string             `json:"slaTier,omitempty"`
	TwoFactorDelete       []*TwoFactorDelete `json:"twoFactorDelete,omitempty"`
}

type GenericDataCentreFields struct {
	ID     string `json:"id,omitempty"`
	Status string `json:"status,omitempty"`

	Name                string `json:"name"`
	Network             string `json:"network"`
	CloudProvider       string `json:"cloudProvider"`
	Region              string `json:"region"`
	ProviderAccountName string `json:"providerAccountName,omitempty"`
	Tags                []*Tag `json:"tags,omitempty"`

	CloudProviderSettings `json:",inline"`
}
