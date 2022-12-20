package v1alpha1

import (
	"encoding/json"
	"net"

	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
	"github.com/instaclustr/operator/pkg/models"
)

// Bundle is deprecated: delete when not used.
type Bundle struct {
	Bundle  string `json:"bundle"`
	Version string `json:"version"`
}

type CloudProviderSettings struct {
	CustomVirtualNetworkID string `json:"customVirtualNetworkId,omitempty"`
	ResourceGroup          string `json:"resourceGroup,omitempty"`
	DiskEncryptionKey      string `json:"diskEncryptionKey,omitempty"`
}

type DataCentre struct {
	// When use one data centre name of field is dataCentreCustomName for APIv1
	Name string `json:"name,omitempty"`

	// Region. APIv1 : "dataCentre"
	Region                string                   `json:"region"`
	CloudProvider         string                   `json:"cloudProvider"`
	ProviderAccountName   string                   `json:"accountName,omitempty"`
	CloudProviderSettings []*CloudProviderSettings `json:"cloudProviderSettings,omitempty"`

	Network  string `json:"network"`
	NodeSize string `json:"nodeSize"`

	// APIv2: replicationFactor; APIv1: numberOfRacks
	RacksNumber int32 `json:"racksNumber"`

	// APIv2: numberOfNodes; APIv1: nodesPerRack.
	NodesNumber int32 `json:"nodesNumber"`

	Tags map[string]string `json:"tags,omitempty"`
}

type DataCentreStatus struct {
	ID              string  `json:"id,omitempty"`
	Status          string  `json:"status,omitempty"`
	Nodes           []*Node `json:"nodes,omitempty"`
	NodeNumber      int32   `json:"nodeNumber,omitempty"`
	EncryptionKeyID string  `json:"encryptionKeyId,omitempty"`
}

type Node struct {
	ID             string   `json:"id,omitempty"`
	Size           string   `json:"size,omitempty"`
	PublicAddress  string   `json:"publicAddress,omitempty"`
	PrivateAddress string   `json:"privateAddress,omitempty"`
	Status         string   `json:"status,omitempty"`
	Roles          []string `json:"roles,omitempty"`
	Rack           string   `json:"rack,omitempty"`
}

type Options struct {
	DataNodeSize                 string `json:"dataNodeSize,omitempty"`
	MasterNodeSize               string `json:"masterNodeSize,omitempty"`
	OpenSearchDashboardsNodeSize string `json:"openSearchDashboardsNodeSize,omitempty"`
}

type Cluster struct {
	// Name [ 3 .. 32 ] characters.
	// APIv2 : "name", APIv1 : "clusterName".
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
	// The PCI compliance standards relate to the security of user data and transactional information.
	// Can only be applied clusters provisioned on AWS_VPC, running Cassandra, Kafka, Elasticsearch and Redis.
	// PCI compliance cannot be enabled if the cluster has Spark.
	//
	// APIv1 : "pciCompliantCluster,omitempty"; APIv2 : pciComplianceMode.
	PCICompliance bool `json:"pciCompliance,omitempty"`
	// Required for APIv2, but for APIv1 set "false" as a default.
	PrivateNetworkCluster bool `json:"privateNetworkCluster,omitempty"`
	// Non-production clusters may receive lower priority support and reduced SLAs.
	// Production tier is not available when using Developer class nodes. See SLA Tier for more information.
	// Enum: "PRODUCTION" "NON_PRODUCTION".
	// Required for APIv2, but for APIv1 set "NON_PRODUCTION" as a default.
	SLATier string `json:"slaTier,omitempty"`

	// APIv2, unlike AP1, receives an array of TwoFactorDelete (<= 1 items);
	TwoFactorDelete []*TwoFactorDelete `json:"twoFactorDelete,omitempty"`
}

type ClusterStatus struct {
	ID                            string              `json:"id,omitempty"`
	Status                        string              `json:"status,omitempty"`
	DataCentres                   []*DataCentreStatus `json:"dataCentres,omitempty"`
	CDCID                         string              `json:"cdcid,omitempty"`
	TwoFactorDeleteEnabled        bool                `json:"twoFactorDeleteEnabled,omitempty"`
	Options                       *Options            `json:"options,omitempty"`
	CurrentClusterOperationStatus string              `json:"currentClusterOperationStatus,omitempty"`
}

type TwoFactorDelete struct {
	// Email address which will be contacted when the cluster is requested to be deleted.
	// APIv1: deleteVerifyEmail; APIv2: confirmationEmail.
	Email string `json:"email"`

	// APIv1: deleteVerifyPhone; APIv2: confirmationPhoneNumber.
	Phone string `json:"phone,omitempty"`
}

type ResizedDataCentre struct {
	CurrentNodeSize      string
	NewNodeSize          string
	DataCentreID         string
	Provider             string
	MasterNewNodeSize    string
	DashboardNewNodeSize string
}

type PatchRequest struct {
	Operation string          `json:"op"`
	Path      string          `json:"path"`
	Value     json.RawMessage `json:"value"`
}

type PrivateLink struct {
	IAMPrincipalARNs []string `json:"iamPrincipalARNs"`
}

func (dc *DataCentre) providerToInstAPIv1() *models.ClusterProvider {
	var instCustomVirtualNetworkId string
	var instResourceGroup string
	var insDiskEncryptionKey string
	if len(dc.CloudProviderSettings) > 0 {
		instCustomVirtualNetworkId = dc.CloudProviderSettings[0].CustomVirtualNetworkID
		instResourceGroup = dc.CloudProviderSettings[0].ResourceGroup
		insDiskEncryptionKey = dc.CloudProviderSettings[0].DiskEncryptionKey
	}

	return &models.ClusterProvider{
		Name:                   dc.CloudProvider,
		AccountName:            dc.ProviderAccountName,
		Tags:                   dc.Tags,
		CustomVirtualNetworkID: instCustomVirtualNetworkId,
		ResourceGroup:          instResourceGroup,
		DiskEncryptionKey:      insDiskEncryptionKey,
	}
}

func (dc *DataCentre) IsNetworkOverlaps(networkToCheck string) (bool, error) {
	_, ipnet, err := net.ParseCIDR(dc.Network)
	if err != nil {
		return false, err
	}

	cassIP, _, err := net.ParseCIDR(networkToCheck)
	if err != nil {
		return false, err
	}

	if ipnet.Contains(cassIP) {
		return true, nil
	}

	return false, nil
}

func (cps *CloudProviderSettings) AWSToInstAPIv2() *modelsv2.AWSSetting {
	return &modelsv2.AWSSetting{
		EBSEncryptionKey:       cps.DiskEncryptionKey,
		CustomVirtualNetworkID: cps.CustomVirtualNetworkID,
	}
}

func (cps *CloudProviderSettings) AzureToInstAPIv2() *modelsv2.AzureSetting {
	return &modelsv2.AzureSetting{
		ResourceGroup: cps.ResourceGroup,
	}
}

func (cps *CloudProviderSettings) GCPToInstAPIv2() *modelsv2.GCPSetting {
	return &modelsv2.GCPSetting{
		CustomVirtualNetworkID: cps.CustomVirtualNetworkID,
	}
}

func (dc DataCentre) TagsToAPIv2() []*modelsv2.Tag {
	var tags []*modelsv2.Tag

	for key, value := range dc.Tags {
		tags = append(tags, &modelsv2.Tag{
			Key:   key,
			Value: value,
		})
	}

	return tags
}

func (tfd *TwoFactorDelete) ToInstAPIv2() *modelsv2.TwoFactorDelete {
	return &modelsv2.TwoFactorDelete{
		ConfirmationPhoneNumber: tfd.Phone,
		ConfirmationEmail:       tfd.Email,
	}
}

func (dc *DataCentre) AreCloudProviderSettingsEqual(
	awsSettings []*modelsv2.AWSSetting,
	gcpSettings []*modelsv2.GCPSetting,
	azureSettings []*modelsv2.AzureSetting) bool {
	switch dc.CloudProvider {
	case modelsv2.AWSVPC:
		if len(dc.CloudProviderSettings) != len(awsSettings) {
			return false
		}

		for i, awsSetting := range awsSettings {
			if dc.CloudProviderSettings[i].DiskEncryptionKey != awsSetting.EBSEncryptionKey ||
				dc.CloudProviderSettings[i].CustomVirtualNetworkID != awsSetting.CustomVirtualNetworkID {
				return false
			}
		}
	case modelsv2.GCP:
		if len(dc.CloudProviderSettings) != len(gcpSettings) {
			return false
		}

		for i, gcpSetting := range gcpSettings {
			if dc.CloudProviderSettings[i].CustomVirtualNetworkID != gcpSetting.CustomVirtualNetworkID {
				return false
			}
		}
	case modelsv2.AZUREAZ:
		if len(dc.CloudProviderSettings) != len(azureSettings) {
			return false
		}

		for i, azureSetting := range azureSettings {
			if dc.CloudProviderSettings[i].ResourceGroup != azureSetting.ResourceGroup {
				return false
			}
		}
	}

	return true
}

func (dc *DataCentre) AreTagsEqual(instTags []*modelsv2.Tag) bool {
	for _, instTag := range instTags {
		if dc.Tags[instTag.Key] != instTag.Value {
			return false
		}
	}

	return true
}

func (c *Cluster) IsTwoFactorDeleteEqual(instTwoFactorDeletes []*modelsv2.TwoFactorDelete) bool {
	if len(c.TwoFactorDelete) != len(instTwoFactorDeletes) {
		return false
	}

	for _, instTFD := range instTwoFactorDeletes {
		tfdFound := false
		for _, k8sTFD := range c.TwoFactorDelete {
			if instTFD.ConfirmationEmail == k8sTFD.Email {
				tfdFound = true

				if instTFD.ConfirmationPhoneNumber != k8sTFD.Phone {
					return false
				}
			}
		}

		if !tfdFound {
			return false
		}
	}

	return true
}

func (dc *DataCentre) SetCloudProviderSettings(instDC *modelsv2.DataCentre) {
	cloudProviderSettings := []*CloudProviderSettings{}
	switch dc.CloudProvider {
	case modelsv2.AWSVPC:
		for _, awsSetting := range instDC.AWSSettings {
			cloudProviderSettings = append(cloudProviderSettings, &CloudProviderSettings{
				CustomVirtualNetworkID: awsSetting.CustomVirtualNetworkID,
				DiskEncryptionKey:      awsSetting.EBSEncryptionKey,
			})
		}
	case modelsv2.GCP:
		for _, gcpSetting := range instDC.GCPSettings {
			cloudProviderSettings = append(cloudProviderSettings, &CloudProviderSettings{
				CustomVirtualNetworkID: gcpSetting.CustomVirtualNetworkID,
			})
		}
	case modelsv2.AZUREAZ:
		for _, azureSetting := range instDC.AzureSettings {
			cloudProviderSettings = append(cloudProviderSettings, &CloudProviderSettings{
				ResourceGroup: azureSetting.ResourceGroup,
			})
		}
	}
	dc.CloudProviderSettings = cloudProviderSettings
}

func (dc *DataCentre) SetCloudProviderSettingsAPIv1(instProviderSettings []*models.ClusterProvider) {
	if len(instProviderSettings) != 0 {
		dc.ProviderAccountName = instProviderSettings[0].AccountName
		dc.CloudProvider = instProviderSettings[0].Name
	}

	cloudProviderSettings := []*CloudProviderSettings{}
	for _, instProviderSetting := range instProviderSettings {
		cloudProviderSettings = append(cloudProviderSettings, &CloudProviderSettings{
			CustomVirtualNetworkID: instProviderSetting.CustomVirtualNetworkID,
			ResourceGroup:          instProviderSetting.ResourceGroup,
			DiskEncryptionKey:      instProviderSetting.DiskEncryptionKey,
		})
	}
	dc.CloudProviderSettings = cloudProviderSettings
}
