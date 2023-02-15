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
	Name                  string                   `json:"name,omitempty"`
	Region                string                   `json:"region"`
	CloudProvider         string                   `json:"cloudProvider"`
	ProviderAccountName   string                   `json:"accountName,omitempty"`
	CloudProviderSettings []*CloudProviderSettings `json:"cloudProviderSettings,omitempty"`
	Network               string                   `json:"network"`
	NodeSize              string                   `json:"nodeSize"`
	NodesNumber           int32                    `json:"nodesNumber"`
	Tags                  map[string]string        `json:"tags,omitempty"`
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
	Name string `json:"name,omitempty"`

	Version string `json:"version,omitempty"`

	// The PCI compliance standards relate to the security of user data and transactional information.
	// Can only be applied clusters provisioned on AWS_VPC, running Cassandra, Kafka, Elasticsearch and Redis.
	// PCI compliance cannot be enabled if the cluster has Spark.
	PCICompliance bool `json:"pciCompliance,omitempty"`

	PrivateNetworkCluster bool `json:"privateNetworkCluster,omitempty"`

	// Non-production clusters may receive lower priority support and reduced SLAs.
	// Production tier is not available when using Developer class nodes. See SLA Tier for more information.
	// Enum: "PRODUCTION" "NON_PRODUCTION".
	SLATier string `json:"slaTier,omitempty"`

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
	MaintenanceEvents             []*MaintenanceEvent `json:"maintenanceEvents,omitempty"`
}

type MaintenanceEvent struct {
	ID                    string `json:"id,omitempty"`
	Description           string `json:"description,omitempty"`
	ScheduledStartTime    string `json:"scheduledStartTime,omitempty"`
	ScheduledEndTime      string `json:"scheduledEndTime,omitempty"`
	ScheduledStartTimeMin string `json:"scheduledStartTimeMin,omitempty"`
	ScheduledStartTimeMax string `json:"scheduledStartTimeMax,omitempty"`
	IsFinalized           bool   `json:"isFinalized,omitempty"`
}

type TwoFactorDelete struct {
	// Email address which will be contacted when the cluster is requested to be deleted.
	Email string `json:"email"`

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

type immutableCluster struct {
	Name                  string
	Version               string
	PCICompliance         bool
	PrivateNetworkCluster bool
	SLATier               string
}

type immutableDC struct {
	Name                string
	Region              string
	CloudProvider       string
	ProviderAccountName string
	Network             string
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

func (dc *DataCentre) TagsToInstAPI(instDC *modelsv2.DataCentre) {
	var tags []*modelsv2.Tag

	for key, value := range dc.Tags {
		tags = append(tags, &modelsv2.Tag{
			Key:   key,
			Value: value,
		})
	}

	instDC.Tags = tags
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

func (dc *DataCentre) SetTagsFromInstAPI(instTags []*modelsv2.Tag) {
	k8sTags := make(map[string]string)
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

func (dcs *DataCentreStatus) SetNodesStatusFromInstAPI(instNodes []*modelsv2.Node) {
	k8sNodes := []*Node{}
	for _, instNode := range instNodes {
		k8sNodes = append(k8sNodes, &Node{
			ID:             instNode.ID,
			Size:           instNode.Size,
			Status:         instNode.Status,
			PublicAddress:  instNode.PublicAddress,
			PrivateAddress: instNode.PrivateAddress,
			Rack:           instNode.Rack,
			Roles:          instNode.Roles,
		})
	}
	dcs.Nodes = k8sNodes
}

func (dcs *DataCentreStatus) AreNodesEqual(instNodes []*modelsv2.Node) bool {
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

func (n *Node) IsNodeEqual(instNode *modelsv2.Node) bool {
	if n == nil || instNode == nil {
		return (n == nil) == (instNode == nil)
	}

	if instNode.ID != n.ID ||
		instNode.Status != n.Status ||
		instNode.Size != n.Size ||
		instNode.Rack != n.Rack ||
		instNode.PrivateAddress != n.PrivateAddress ||
		instNode.PublicAddress != n.PublicAddress ||
		!slices.Equal(instNode.Roles, n.Roles) {
		return false
	}

	return true
}

func (dc *DataCentre) SetDefaultValues() {
	if dc.ProviderAccountName == "" {
		dc.ProviderAccountName = models.DefaultAccountName
	}

	if len(dc.CloudProviderSettings) == 0 {
		dc.CloudProviderSettings = append(dc.CloudProviderSettings, &CloudProviderSettings{
			CustomVirtualNetworkID: "",
			ResourceGroup:          "",
			DiskEncryptionKey:      "",
		})
	}
}

func (c *Cluster) newImmutableFields() immutableCluster {
	return immutableCluster{
		Name:                  c.Name,
		Version:               c.Version,
		PCICompliance:         c.PCICompliance,
		PrivateNetworkCluster: c.PrivateNetworkCluster,
		SLATier:               c.SLATier,
	}
}

func (c *ClusterStatus) AreMaintenanceEventsEqual(instEvents []*MaintenanceEvent) bool {
	if len(c.MaintenanceEvents) != len(instEvents) {
		return false
	}

	for _, instEvent := range instEvents {
		for _, k8sEvent := range c.MaintenanceEvents {
			if instEvent.ID == k8sEvent.ID {
				if *instEvent != *k8sEvent {
					return false
				}

				break
			}
		}
	}

	return true
}
