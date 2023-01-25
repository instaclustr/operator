package v1alpha1

import (
	"encoding/json"
	"net"

	"k8s.io/utils/strings/slices"

	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
	"github.com/instaclustr/operator/pkg/models"
)

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

func (dc *DataCentre) CloudProviderSettingsToInstAPI(instDC *modelsv2.DataCentre) {
	switch dc.CloudProvider {
	case modelsv2.AWSVPC:
		awsSettings := []*modelsv2.AWSSetting{}
		for _, providerSettings := range dc.CloudProviderSettings {
			awsSettings = append(awsSettings, providerSettings.AWSToInstAPI())
		}
		instDC.AWSSettings = awsSettings
	case modelsv2.AZURE, modelsv2.AZUREAZ:
		azureSettings := []*modelsv2.AzureSetting{}
		for _, providerSettings := range dc.CloudProviderSettings {
			azureSettings = append(azureSettings, providerSettings.AzureToInstAPI())
			instDC.AzureSettings = azureSettings
		}
	case modelsv2.GCP:
		gcpSettings := []*modelsv2.GCPSetting{}
		for _, providerSettings := range dc.CloudProviderSettings {
			gcpSettings = append(gcpSettings, providerSettings.GCPToInstAPI())
		}
		instDC.GCPSettings = gcpSettings
	}

}

func (cps *CloudProviderSettings) AWSToInstAPI() *modelsv2.AWSSetting {
	return &modelsv2.AWSSetting{
		EBSEncryptionKey:       cps.DiskEncryptionKey,
		CustomVirtualNetworkID: cps.CustomVirtualNetworkID,
	}
}

func (cps *CloudProviderSettings) AzureToInstAPI() *modelsv2.AzureSetting {
	return &modelsv2.AzureSetting{
		ResourceGroup: cps.ResourceGroup,
	}
}

func (cps *CloudProviderSettings) GCPToInstAPI() *modelsv2.GCPSetting {
	return &modelsv2.GCPSetting{
		CustomVirtualNetworkID: cps.CustomVirtualNetworkID,
	}
}

func (dc DataCentre) TagsToInstAPI(instDC *modelsv2.DataCentre) {
	var tags []*modelsv2.Tag

	for key, value := range dc.Tags {
		tags = append(tags, &modelsv2.Tag{
			Key:   key,
			Value: value,
		})
	}

	instDC.Tags = tags
}

func (tfd *TwoFactorDelete) ToInstAPI() *modelsv2.TwoFactorDelete {
	return &modelsv2.TwoFactorDelete{
		ConfirmationPhoneNumber: tfd.Phone,
		ConfirmationEmail:       tfd.Email,
	}
}

func (c *Cluster) TwoFactorDeletesToInstAPI() []*modelsv2.TwoFactorDelete {
	var TFDs []*modelsv2.TwoFactorDelete
	for _, k8sTFD := range c.TwoFactorDelete {
		TFDs = append(TFDs, k8sTFD.ToInstAPI())
	}
	return TFDs
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
	case modelsv2.AZURE, modelsv2.AZUREAZ:
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

func (dc *DataCentre) SetCloudProviderSettingsFromInstAPI(instDC *modelsv2.DataCentre) {
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
	case modelsv2.AZURE, modelsv2.AZUREAZ:
		for _, azureSetting := range instDC.AzureSettings {
			cloudProviderSettings = append(cloudProviderSettings, &CloudProviderSettings{
				ResourceGroup: azureSetting.ResourceGroup,
			})
		}
	}
	dc.CloudProviderSettings = cloudProviderSettings
}

func (c *Cluster) SetTwoFactorDeletesFromInstAPI(instTFDs []*modelsv2.TwoFactorDelete) {
	var k8sTFD []*TwoFactorDelete
	for _, instTFD := range instTFDs {
		k8sTFD = append(k8sTFD, &TwoFactorDelete{
			Email: instTFD.ConfirmationEmail,
			Phone: instTFD.ConfirmationPhoneNumber,
		})
	}
	c.TwoFactorDelete = k8sTFD
}

func (dc *DataCentre) SetTagsFromInstAPI(instTags []*modelsv2.Tag) {
	var k8sTags map[string]string
	for _, instTag := range instTags {
		k8sTags[instTag.Key] = instTag.Value
	}
	dc.Tags = k8sTags
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

func (dcs *DataCentreStatus) SetNodesStatusFromInstAPI(instNodes []*models.NodeStatusV2) {
	k8sNodes := []*Node{}
	for _, instNode := range instNodes {
		k8sNodes = append(k8sNodes, &Node{
			ID:             instNode.ID,
			Size:           instNode.NodeSize,
			Status:         instNode.Status,
			PublicAddress:  instNode.PublicAddress,
			PrivateAddress: instNode.PrivateAddress,
			Rack:           instNode.Rack,
			Roles:          instNode.NodeRoles,
		})
	}
	dcs.Nodes = k8sNodes
}

func (dcs *DataCentreStatus) AreNodesEqual(instNodes []*models.NodeStatusV2) bool {
	if len(dcs.Nodes) != len(instNodes) {
		return false
	}

	for _, instNode := range instNodes {
		for _, k8sNode := range dcs.Nodes {
			if instNode.ID == k8sNode.ID {
				if !k8sNode.IsNodeEqual(instNode) {
					return false
				}

				break
			}
		}
	}

	return true
}

func (n *Node) IsNodeEqual(instNode *models.NodeStatusV2) bool {
	if n == nil || instNode == nil {
		if (n == nil) != (instNode == nil) {
			return false
		}

		return true
	}

	if instNode.ID != n.ID ||
		instNode.Status != n.Status ||
		instNode.NodeSize != n.Size ||
		instNode.Rack != n.Rack ||
		instNode.PrivateAddress != n.PrivateAddress ||
		instNode.PublicAddress != n.PublicAddress ||
		!slices.Equal(instNode.NodeRoles, n.Roles) {
		return false
	}

	return true
}
