package models

import (
	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
)

const (
	AWSAccessKeyID     = "awsAccessKeyId"
	AWSSecretAccessKey = "awsSecretAccessKey"

	SharedProvisioningType   = "SHARED"
	PackagedProvisioningType = "PACKAGED"
	StandardProvisioningType = "STANDARD"
)

type CadenceSpec struct {
	ID                            string                         `json:"id,omitempty"`
	Name                          string                         `json:"name"`
	CadenceVersion                string                         `json:"cadenceVersion"`
	DataCentres                   []*CadenceDataCentre           `json:"dataCentres"`
	SharedProvisioning            []*CadenceSharedProvisioning   `json:"sharedProvisioning,omitempty"`
	StandardProvisioning          []*CadenceStandardProvisioning `json:"standardProvisioning,omitempty"`
	PCIComplianceMode             bool                           `json:"pciComplianceMode"`
	TwoFactorDelete               []*modelsv2.TwoFactorDelete    `json:"twoFactorDelete,omitempty"`
	UseCadenceWebAuth             bool                           `json:"useCadenceWebAuth"`
	PrivateNetworkCluster         bool                           `json:"privateNetworkCluster"`
	SLATier                       string                         `json:"slaTier"`
	AWSArchival                   []*AWSArchival                 `json:"awsArchival,omitempty"`
	CurrentClusterOperationStatus string                         `json:"currentClusterOperationStatus,omitempty"`
	Status                        string                         `json:"status,omitempty"`
}

type CadenceDataCentre struct {
	modelsv2.DataCentre       `json:",inline"`
	ClientToClusterEncryption bool             `json:"clientToClusterEncryption"`
	ID                        string           `json:"id,omitempty"`
	Nodes                     []*modelsv2.Node `json:"nodes,omitempty"`
	Status                    string           `json:"status,omitempty"`
	PrivateLink               []*PrivateLink   `json:"privateLink,omitempty"`
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
