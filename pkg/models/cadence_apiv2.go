/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package models

const (
	AWSAccessKeyID     = "awsAccessKeyId"
	AWSSecretAccessKey = "awsSecretAccessKey"

	SharedProvisioningType   = "SHARED"
	PackagedProvisioningType = "PACKAGED"
	StandardProvisioningType = "STANDARD"
)

type CadenceCluster struct {
	ClusterStatus          `json:",inline"`
	Name                   string                         `json:"name"`
	CadenceVersion         string                         `json:"cadenceVersion"`
	DataCentres            []*CadenceDataCentre           `json:"dataCentres"`
	SharedProvisioning     []*CadenceSharedProvisioning   `json:"sharedProvisioning,omitempty"`
	StandardProvisioning   []*CadenceStandardProvisioning `json:"standardProvisioning,omitempty"`
	PCIComplianceMode      bool                           `json:"pciComplianceMode"`
	TwoFactorDelete        []*TwoFactorDelete             `json:"twoFactorDelete,omitempty"`
	UseCadenceWebAuth      bool                           `json:"useCadenceWebAuth"`
	PrivateNetworkCluster  bool                           `json:"privateNetworkCluster"`
	SLATier                string                         `json:"slaTier"`
	AWSArchival            []*AWSArchival                 `json:"awsArchival,omitempty"`
	TargetPrimaryCadence   []*TargetCadence               `json:"targetPrimaryCadence,omitempty"`
	TargetSecondaryCadence []*TargetCadence               `json:"targetSecondaryCadence,omitempty"`
	ResizeSettings         []*ResizeSettings              `json:"resizeSettings,omitempty"`
	Description            string                         `json:"description,omitempty"`
}

type CadenceDataCentre struct {
	DataCentre                `json:",inline"`
	ClientToClusterEncryption bool           `json:"clientToClusterEncryption"`
	PrivateLink               []*PrivateLink `json:"privateLink,omitempty"`
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

type TargetCadence struct {
	DependencyCDCID   string `json:"dependencyCdcId"`
	DependencyVPCType string `json:"dependencyVpcType"`
}

type AWSArchival struct {
	ArchivalS3Region   string `json:"archivalS3Region,omitempty"`
	AWSAccessKeyID     string `json:"awsAccessKeyId,omitempty"`
	ArchivalS3URI      string `json:"archivalS3Uri,omitempty"`
	AWSSecretAccessKey string `json:"awsSecretAccessKey,omitempty"`
}

type CadenceClusterAPIUpdate struct {
	DataCentres []*CadenceDataCentre `json:"dataCentres"`
}
