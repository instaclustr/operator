package models

import modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"

type CadenceClusterAPIv2 struct {
	Name                  string                         `json:"name"`
	CadenceVersion        string                         `json:"cadenceVersion"`
	CadenceDataCentres    []*CadenceDataCentre           `json:"dataCentres"`
	SharedProvisioning    []*CadenceSharedProvisioning   `json:"sharedProvisioning,omitempty"`
	StandardProvisioning  []*CadenceStandardProvisioning `json:"standardProvisioning,omitempty"`
	PCIComplianceMode     bool                           `json:"pciComplianceMode"`
	TwoFactorDelete       []*modelsv2.TwoFactorDelete    `json:"twoFactorDelete,omitempty"`
	UseCadenceWebAuth     bool                           `json:"useCadenceWebAuth"`
	PrivateNetworkCluster bool                           `json:"privateNetworkCluster"`
	SLATier               string                         `json:"slaTier"`
	AWSArchival           *AWSArchival                   `json:"awsArchival,omitempty"`
}

type CadenceDataCentre struct {
	ClientToClusterEncryption bool                     `json:"clientToClusterEncryption"`
	CloudProvider             string                   `json:"cloudProvider"`
	Name                      string                   `json:"name"`
	Network                   string                   `json:"network"`
	NodeSize                  string                   `json:"nodeSize"`
	NumberOfNodes             int                      `json:"numberOfNodes"`
	Region                    string                   `json:"region"`
	AWSSettings               []*modelsv2.AWSSetting   `json:"awsSettings,omitempty"`
	GCPSetting                []*modelsv2.GCPSetting   `json:"gcpSettings,omitempty"`
	AzureSettings             []*modelsv2.AzureSetting `json:"azureSettings,omitempty"`
	ProviderAccountName       string                   `json:"providerAccountName,omitempty"`
	Tags                      []*modelsv2.Tag          `json:"tags,omitempty"`
}

type CadenceSharedProvisioning struct {
	UseAdvancedVisibility bool `json:"useAdvancedVisibility"`
}

type CadenceStandardProvisioning struct {
	AdvancedVisibility []*AdvancedVisibility `json:"advancedVisibility,omitempty"`
	TargetCassandra    *TargetCassandra      `json:"targetCassandra"`
}

type AdvancedVisibility struct {
	TargetKafka      *TargetKafka      `json:"targetKafka"`
	TargetOpenSearch *TargetOpenSearch `json:"targetOpenSearch"`
}

type TargetKafka struct {
	DependencyCDCID   string `json:"dependencyCdcId"`
	DependencyVPCType string `json:"dependencyVpcType"`
}

type TargetOpenSearch struct {
	DependencyCDCID   string `json:"dependencyCdcId"`
	DependencyVPCType string `json:"dependencyVpcType"`
}

type TargetCassandra struct {
	DependencyCDCID   string `json:"dependencyCdcId"`
	DependencyVPCType string `json:"dependencyVpcType"`
}

type AWSArchival struct {
	ArchivalS3Region   string `json:"archivalS3Region,omitempty"`
	AWSAccessKeyID     string `json:"awsAccessKeyId,omitempty"`
	ArchivalS3URI      string `json:"archivalS3Uri,omitempty"`
	AWSSecretAccessKey string `json:"awsSecretAccessKey,omitempty"`
}
