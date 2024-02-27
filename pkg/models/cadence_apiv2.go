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

// Related to Packaged Provisioning
const (
	DeveloperSize         = "Developer"
	ProductionStarterSize = "Production-Starter"
	ProductionSmallSize   = "Production-Small"
)

var (
	AWSCassandraSizes = map[string]string{
		DeveloperSize:         "CAS-DEV-t4g.small-5",
		ProductionStarterSize: "CAS-PRD-m6g.large-120",
		ProductionSmallSize:   "CAS-PRD-r6g.large-800",
	}
	AzureCassandraSizes = map[string]string{
		DeveloperSize:         "Standard_DS2_v2-256-an",
		ProductionStarterSize: "Standard_DS2_v2-256-an",
		ProductionSmallSize:   "Standard_DS12_v2-512-an",
	}
	GCPCassandraSizes = map[string]string{
		DeveloperSize:         "CAS-DEV-n1-standard-1-5",
		ProductionStarterSize: "CAS-PRD-n2-standard-2-250",
		ProductionSmallSize:   "CAS-PRD-n2-highmem-2-400",
	}
	AWSKafkaSizes = map[string]string{
		DeveloperSize:         "KFK-DEV-t4g.small-5",
		ProductionStarterSize: "KFK-PRD-r6g.large-250",
		ProductionSmallSize:   "KFK-PRD-r6g.large-400",
	}
	AzureKafkaSizes = map[string]string{
		DeveloperSize:         "Standard_DS1_v2-32",
		ProductionStarterSize: "Standard_DS11_v2-512",
		ProductionSmallSize:   "Standard_DS11_v2-512",
	}
	GCPKafkaSizes = map[string]string{
		DeveloperSize:         "n1-standard-1-80",
		ProductionStarterSize: "n1-highmem-2-400",
		ProductionSmallSize:   "n1-highmem-2-400",
	}
	AWSOpenSearchSizes = map[string]string{
		DeveloperSize:         "SRH-DEV-t4g.small-5",
		ProductionStarterSize: "SRH-PRD-m6g.large-250",
		ProductionSmallSize:   "SRH-PRD-r6g.large-800",
	}
	AzureOpenSearchSizes = map[string]string{
		DeveloperSize:         "SRH-DEV-DS1_v2-5-an",
		ProductionStarterSize: "SRH-PRD-D2s_v5-250-an",
		ProductionSmallSize:   "SRH-PRD-E2s_v4-800-an",
	}
	GCPOpenSearchSizes = map[string]string{
		DeveloperSize:         "SRH-DEV-n1-standard-1-30",
		ProductionStarterSize: "SRH-PRD-n2-standard-2-250",
		ProductionSmallSize:   "SRH-PRD-n2-highmem-2-800",
	}
)

type solutionSizesMap map[string]map[string]map[string]string

var SolutionSizesMap = solutionSizesMap{
	AWSVPC: {
		CassandraAppKind:  AWSCassandraSizes,
		KafkaAppKind:      AWSKafkaSizes,
		OpenSearchAppKind: AWSOpenSearchSizes,
	},
	AZUREAZ: {
		CassandraAppKind:  AzureCassandraSizes,
		KafkaAppKind:      AzureKafkaSizes,
		OpenSearchAppKind: AzureOpenSearchSizes,
	},
	GCP: {
		CassandraAppKind:  GCPCassandraSizes,
		KafkaAppKind:      GCPKafkaSizes,
		OpenSearchAppKind: GCPOpenSearchSizes,
	},
}

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
