package v1alpha1

import (
	"encoding/json"
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
	Name    string `json:"name"`
	Version string `json:"version"`
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
	SLATier string `json:"slaTier"`

	// APIv2, unlike AP1, receives an array of TwoFactorDelete (<= 1 items);
	TwoFactorDelete []*TwoFactorDelete `json:"twoFactorDelete,omitempty"`
}

type ClusterStatus struct {
	ID                     string              `json:"id,omitempty"`
	Status                 string              `json:"status,omitempty"`
	DataCentres            []*DataCentreStatus `json:"dataCentres,omitempty"`
	CDCID                  string              `json:"cdcid,omitempty"`
	TwoFactorDeleteEnabled bool                `json:"twoFactorDeleteEnabled,omitempty"`
	Options                *Options            `json:"options,omitempty"`
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

func (dataCentre *DataCentre) providerToInstAPIv1() *models.ClusterProvider {
	var instCustomVirtualNetworkId string
	var instResourceGroup string
	var insDiskEncryptionKey string
	if len(dataCentre.CloudProviderSettings) > 0 {
		instCustomVirtualNetworkId = dataCentre.CloudProviderSettings[0].CustomVirtualNetworkID
		instResourceGroup = dataCentre.CloudProviderSettings[0].ResourceGroup
		insDiskEncryptionKey = dataCentre.CloudProviderSettings[0].DiskEncryptionKey
	}

	return &models.ClusterProvider{
		Name:                   dataCentre.CloudProvider,
		AccountName:            dataCentre.ProviderAccountName,
		Tags:                   dataCentre.Tags,
		CustomVirtualNetworkId: instCustomVirtualNetworkId,
		ResourceGroup:          instResourceGroup,
		DiskEncryptionKey:      insDiskEncryptionKey,
	}
}
