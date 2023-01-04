package models

const (
	Cadence            = "CADENCE"
	CadenceNodePurpose = "CADENCE"

	AWSAccessKeyID     = "awsAccessKeyId"
	AWSSecretAccessKey = "awsSecretAccessKey"

	SharedProvisioningType   = "SHARED"
	PackagedProvisioningType = "PACKAGED"
	StandardProvisioningType = "STANDARD"
)

type CadenceClusterAPIv1 struct {
	Cluster     `json:",inline"`
	Bundles     []*CadenceBundleAPIv1     `json:"bundles"`
	DataCentres []*CadenceDataCentreAPIv1 `json:"dataCentres,omitempty"`
}

type CadenceDataCentreAPIv1 struct {
	DataCentre `json:",inline"`
	Bundles    []*CadenceBundleAPIv1 `json:"bundles"`
}

type CadenceBundleAPIv1 struct {
	Bundle  `json:",inline"`
	Options *CadenceBundleOptionsAPIv1 `json:"options"`
}

type CadenceBundleOptionsAPIv1 struct {
	UseAdvancedVisibility   bool   `json:"useAdvancedVisibility,omitempty"`
	UseCadenceWebAuth       bool   `json:"useCadenceWebAuth,omitempty"`
	ClientEncryption        bool   `json:"clientEncryption,omitempty"`
	TargetCassandraCDCID    string `json:"targetCassandraCdcId"`
	TargetCassandraVPCType  string `json:"targetCassandraVpcType,omitempty"`
	TargetKafkaCDCID        string `json:"targetKafkaCdcId,omitempty"`
	TargetKafkaVPCType      string `json:"targetKafkaVpcType,omitempty"`
	TargetOpenSearchCDCID   string `json:"targetOpenSearchCdcId,omitempty"`
	TargetOpenSearchVPCType string `json:"targetOpenSearchVpcType,omitempty"`
	EnableArchival          bool   `json:"enableArchival,omitempty"`
	ArchivalS3URI           string `json:"archivalS3Uri,omitempty"`
	ArchivalS3Region        string `json:"archivalS3Region,omitempty"`
	AWSAccessKeyID          string `json:"awsAccessKeyId,omitempty"`
	AWSSecretAccessKey      string `json:"awsSecretAccessKey,omitempty"`
	CadenceNodeCount        int    `json:"cadenceNodeCount"`
	ProvisioningType        string `json:"provisioningType"`
}

type CadenceUpdatedFields struct {
	NodeSizeUpdated        bool `json:"nodeSize"`
	DescriptionUpdated     bool `json:"description"`
	TwoFactorDeleteUpdated bool `json:"twoFactorDelete"`
}
