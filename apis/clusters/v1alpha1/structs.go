package v1alpha1

import (
	"encoding/json"
	"net"

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
	NodesNumber           int                      `json:"nodesNumber"`
	Tags                  map[string]string        `json:"tags,omitempty"`
}

type DataCentreStatus struct {
	ID              string  `json:"id,omitempty"`
	Status          string  `json:"status,omitempty"`
	Nodes           []*Node `json:"nodes,omitempty"`
	NodeNumber      int     `json:"nodeNumber,omitempty"`
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
	State                         string              `json:"state,omitempty"`
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

type PatchRequest struct {
	Operation string          `json:"op"`
	Path      string          `json:"path"`
	Value     json.RawMessage `json:"value"`
}

type PrivateLink struct {
	AdvertisedHostname string `json:"advertisedHostname"`
}

type PrivateLinkV1 struct {
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

func (c *Cluster) IsEqual(cluster Cluster) bool {
	return c.Name == cluster.Name &&
		c.Version == cluster.Version &&
		c.PCICompliance == cluster.PCICompliance &&
		c.PrivateNetworkCluster == cluster.PrivateNetworkCluster &&
		c.SLATier == cluster.SLATier &&
		c.IsTwoFactorDeleteEqual(cluster.TwoFactorDelete)
}

func (c *Cluster) IsTwoFactorDeleteEqual(tfds []*TwoFactorDelete) bool {
	if len(c.TwoFactorDelete) != len(tfds) {
		return false
	}

	for i, tfd := range tfds {
		if *tfd != *c.TwoFactorDelete[i] {
			return false
		}
	}

	return true
}

func (tfd *TwoFactorDelete) ToInstAPI() *models.TwoFactorDelete {
	return &models.TwoFactorDelete{
		ConfirmationPhoneNumber: tfd.Phone,
		ConfirmationEmail:       tfd.Email,
	}
}

func (c *Cluster) TwoFactorDeletesToInstAPI() (TFDs []*models.TwoFactorDelete) {
	for _, tfd := range c.TwoFactorDelete {
		TFDs = append(TFDs, tfd.ToInstAPI())
	}
	return
}

func (c *Cluster) TwoFactorDeleteToInstAPIv1() *models.TwoFactorDeleteV1 {
	if len(c.TwoFactorDelete) == 0 {
		return nil
	}

	return &models.TwoFactorDeleteV1{
		DeleteVerifyEmail: c.TwoFactorDelete[0].Email,
		DeleteVerifyPhone: c.TwoFactorDelete[0].Phone,
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

func (dc *DataCentre) ToInstAPI() models.DataCentre {
	providerSettings := dc.CloudProviderSettingsToInstAPI()
	return models.DataCentre{
		Name:                dc.Name,
		Network:             dc.Network,
		NodeSize:            dc.NodeSize,
		NumberOfNodes:       dc.NodesNumber,
		AWSSettings:         providerSettings.AWSSettings,
		GCPSettings:         providerSettings.GCPSettings,
		AzureSettings:       providerSettings.AzureSettings,
		Tags:                dc.TagsToInstAPI(),
		CloudProvider:       dc.CloudProvider,
		Region:              dc.Region,
		ProviderAccountName: dc.ProviderAccountName,
	}
}

func (dc *DataCentre) CloudProviderSettingsToInstAPI() *models.CloudProviderSettings {
	iSettings := &models.CloudProviderSettings{}
	switch dc.CloudProvider {
	case models.AWSVPC:
		awsSettings := []*models.AWSSetting{}
		for _, providerSettings := range dc.CloudProviderSettings {
			awsSettings = append(awsSettings, providerSettings.AWSToInstAPI())
		}
		iSettings.AWSSettings = awsSettings
	case models.AZURE, models.AZUREAZ:
		azureSettings := []*models.AzureSetting{}
		for _, providerSettings := range dc.CloudProviderSettings {
			azureSettings = append(azureSettings, providerSettings.AzureToInstAPI())
		}
		iSettings.AzureSettings = azureSettings
	case models.GCP:
		gcpSettings := []*models.GCPSetting{}
		for _, providerSettings := range dc.CloudProviderSettings {
			gcpSettings = append(gcpSettings, providerSettings.GCPToInstAPI())
		}
		iSettings.GCPSettings = gcpSettings
	}

	return iSettings
}

func (cps *CloudProviderSettings) AWSToInstAPI() *models.AWSSetting {
	return &models.AWSSetting{
		EBSEncryptionKey:       cps.DiskEncryptionKey,
		CustomVirtualNetworkID: cps.CustomVirtualNetworkID,
	}
}

func (cps *CloudProviderSettings) AzureToInstAPI() *models.AzureSetting {
	return &models.AzureSetting{
		ResourceGroup: cps.ResourceGroup,
	}
}

func (cps *CloudProviderSettings) GCPToInstAPI() *models.GCPSetting {
	return &models.GCPSetting{
		CustomVirtualNetworkID: cps.CustomVirtualNetworkID,
	}
}

func (dc *DataCentre) TagsToInstAPI() (tags []*models.Tag) {
	for key, value := range dc.Tags {
		tags = append(tags, &models.Tag{
			Key:   key,
			Value: value,
		})
	}

	return
}

func (dc *DataCentre) IsEqual(iDC DataCentre) bool {
	return iDC.Name == dc.Name &&
		iDC.Region == dc.Region &&
		iDC.CloudProvider == dc.CloudProvider &&
		iDC.ProviderAccountName == dc.ProviderAccountName &&
		dc.AreCloudProviderSettingsEqual(iDC.CloudProviderSettings) &&
		iDC.Network == dc.Network &&
		iDC.NodeSize == dc.NodeSize &&
		iDC.NodesNumber == dc.NodesNumber &&
		dc.AreTagsEqual(iDC.Tags)
}

func (dc *DataCentre) AreCloudProviderSettingsEqual(settings []*CloudProviderSettings) bool {
	if len(dc.CloudProviderSettings) != len(settings) {
		return false
	}

	for i, setting := range settings {
		if *dc.CloudProviderSettings[i] != *setting {
			return false
		}
	}

	return true
}

func (dc *DataCentre) AreTagsEqual(tags map[string]string) bool {
	if len(dc.Tags) != len(tags) {
		return false
	}

	for key, val := range tags {
		if value, exists := dc.Tags[key]; !exists || value != val {
			return false
		}
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

func (cs *ClusterStatus) AreMaintenanceEventsEqual(iEvents []*MaintenanceEvent) bool {
	if len(cs.MaintenanceEvents) != len(iEvents) {
		return false
	}

	for _, iEvent := range iEvents {
		for _, k8sEvent := range cs.MaintenanceEvents {
			if iEvent.ID == k8sEvent.ID {
				if *iEvent != *k8sEvent {
					return false
				}

				break
			}
		}
	}

	return true
}

func (cs *ClusterStatus) DCFromInstAPI(iDC models.DataCentre) *DataCentreStatus {
	return &DataCentreStatus{
		ID:         iDC.ID,
		Status:     iDC.Status,
		Nodes:      cs.NodesFromInstAPI(iDC.Nodes),
		NodeNumber: iDC.NumberOfNodes,
	}
}

func (c *Cluster) TwoFactorDeleteFromInstAPI(iTFDs []*models.TwoFactorDelete) (tfd []*TwoFactorDelete) {
	for _, iTFD := range iTFDs {
		tfd = append(tfd, &TwoFactorDelete{
			Email: iTFD.ConfirmationEmail,
			Phone: iTFD.ConfirmationPhoneNumber,
		})
	}
	return
}

func (c *Cluster) DCFromInstAPI(iDC models.DataCentre) DataCentre {
	return DataCentre{
		Name:                  iDC.Name,
		Region:                iDC.Region,
		CloudProvider:         iDC.CloudProvider,
		ProviderAccountName:   iDC.ProviderAccountName,
		CloudProviderSettings: c.CloudProviderSettingsFromInstAPI(iDC),
		Network:               iDC.Network,
		NodeSize:              iDC.NodeSize,
		NodesNumber:           iDC.NumberOfNodes,
		Tags:                  c.TagsFromInstAPI(iDC.Tags),
	}
}

func (c *Cluster) TagsFromInstAPI(iTags []*models.Tag) map[string]string {
	newTags := map[string]string{}
	for _, iTag := range iTags {
		newTags[iTag.Key] = iTag.Value
	}
	return newTags
}

func (c *Cluster) CloudProviderSettingsFromInstAPI(iDC models.DataCentre) (settings []*CloudProviderSettings) {
	switch iDC.CloudProvider {
	case models.AWSVPC:
		for _, awsSetting := range iDC.AWSSettings {
			settings = append(settings, &CloudProviderSettings{
				CustomVirtualNetworkID: awsSetting.CustomVirtualNetworkID,
				DiskEncryptionKey:      awsSetting.EBSEncryptionKey,
			})
		}
	case models.GCP:
		for _, gcpSetting := range iDC.GCPSettings {
			settings = append(settings, &CloudProviderSettings{
				CustomVirtualNetworkID: gcpSetting.CustomVirtualNetworkID,
			})
		}
	case models.AZURE, models.AZUREAZ:
		for _, azureSetting := range iDC.AzureSettings {
			settings = append(settings, &CloudProviderSettings{
				ResourceGroup: azureSetting.ResourceGroup,
			})
		}
	}
	return
}

func (c *Cluster) CloudProviderSettingsFromInstAPIv1(iProviders []*models.ClusterProviderV1) (accountName string, settings []*CloudProviderSettings) {
	for _, iProvider := range iProviders {
		accountName = iProvider.AccountName
		switch iProvider.Name {
		case models.AWSVPC:
			settings = append(settings, &CloudProviderSettings{
				CustomVirtualNetworkID: iProvider.CustomVirtualNetworkID,
				DiskEncryptionKey:      iProvider.DiskEncryptionKey,
			})
		case models.GCP:
			settings = append(settings, &CloudProviderSettings{
				CustomVirtualNetworkID: iProvider.CustomVirtualNetworkID,
			})
		case models.AZURE, models.AZUREAZ:
			settings = append(settings, &CloudProviderSettings{
				ResourceGroup: iProvider.ResourceGroup,
			})
		}
	}
	return
}

func (cs *ClusterStatus) NodesFromInstAPI(iNodes []*models.Node) (nodes []*Node) {
	for _, iNode := range iNodes {
		nodes = append(nodes, &Node{
			ID:             iNode.ID,
			Size:           iNode.Size,
			PublicAddress:  iNode.PublicAddress,
			PrivateAddress: iNode.PrivateAddress,
			Status:         iNode.Status,
			Roles:          iNode.Roles,
			Rack:           iNode.Rack,
		})
	}
	return nodes
}

func (cs *ClusterStatus) NodesFromInstAPIv1(iNodes []*models.NodeStatusV1) (nodes []*Node) {
	for _, iNode := range iNodes {
		nodes = append(nodes, &Node{
			ID:             iNode.ID,
			Size:           iNode.Size,
			PublicAddress:  iNode.PublicAddress,
			PrivateAddress: iNode.PrivateAddress,
			Status:         iNode.NodeStatus,
			Rack:           iNode.Rack,
		})
	}
	return nodes
}

func arePrivateLinksEqual(a, b []*PrivateLink) bool {
	if len(a) != len(b) {
		return false
	}

	for i, privateLink := range a {
		if *b[i] != *privateLink {
			return false
		}
	}

	return true
}
