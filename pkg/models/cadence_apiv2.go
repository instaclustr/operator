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
)

type CadenceCluster struct {
	GenericClusterFields `json:",inline"`

	DataCentres            []*CadenceDataCentre           `json:"dataCentres"`
	SharedProvisioning     []*CadenceSharedProvisioning   `json:"sharedProvisioning,omitempty"`
	StandardProvisioning   []*CadenceStandardProvisioning `json:"standardProvisioning,omitempty"`
	AWSArchival            []*AWSArchival                 `json:"awsArchival,omitempty"`
	TwoFactorDelete        []*TwoFactorDelete             `json:"twoFactorDelete,omitempty"`
	TargetPrimaryCadence   []*CadenceDependencyTarget     `json:"targetPrimaryCadence,omitempty"`
	TargetSecondaryCadence []*CadenceDependencyTarget     `json:"targetSecondaryCadence,omitempty"`
	ResizeSettings         []*ResizeSettings              `json:"resizeSettings,omitempty"`

	CadenceVersion    string `json:"cadenceVersion"`
	UseHTTPAPI        bool   `json:"useHttpApi,omitempty"`
	PCIComplianceMode bool   `json:"pciComplianceMode"`
	UseCadenceWebAuth bool   `json:"useCadenceWebAuth"`
}

type CadenceDataCentre struct {
	GenericDataCentreFields `json:",inline"`

	ClientToClusterEncryption bool   `json:"clientToClusterEncryption"`
	NumberOfNodes             int    `json:"numberOfNodes"`
	NodeSize                  string `json:"nodeSize"`

	PrivateLink []*PrivateLink `json:"privateLink,omitempty"`
	Nodes       []*Node        `json:"nodes"`
}

type CadenceSharedProvisioning struct {
	UseAdvancedVisibility bool `json:"useAdvancedVisibility"`
}

type CadenceStandardProvisioning struct {
	AdvancedVisibility []*AdvancedVisibility    `json:"advancedVisibility,omitempty"`
	TargetCassandra    *CadenceDependencyTarget `json:"targetCassandra"`
}

type AdvancedVisibility struct {
	TargetKafka      *CadenceDependencyTarget `json:"targetKafka"`
	TargetOpenSearch *CadenceDependencyTarget `json:"targetOpenSearch"`
}

type CadenceDependencyTarget struct {
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
	DataCentres    []*CadenceDataCentre `json:"dataCentres"`
	ResizeSettings []*ResizeSettings    `json:"resizeSettings,omitempty"`
}
