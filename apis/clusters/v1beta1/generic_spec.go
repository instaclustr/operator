package v1beta1

import (
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/utils/slices"
)

type GenericClusterSpec struct {
	// Name [ 3 .. 32 ] characters.
	Name string `json:"name,omitempty"`

	Version string `json:"version,omitempty"`

	PrivateNetwork bool `json:"privateNetwork,omitempty"`

	// Non-production clusters may receive lower priority support and reduced SLAs.
	// Production tier is not available when using Developer class nodes. See SLA Tier for more information.
	// Enum: "PRODUCTION" "NON_PRODUCTION".
	SLATier string `json:"slaTier,omitempty"`

	Description string `json:"description,omitempty"`

	TwoFactorDelete []*TwoFactorDelete `json:"twoFactorDelete,omitempty"`
}

func (s *GenericClusterSpec) Equals(o *GenericClusterSpec) bool {
	return s.Name == o.Name &&
		s.Version == o.Version &&
		s.PrivateNetwork == o.PrivateNetwork &&
		s.SLATier == o.SLATier &&
		s.Description == o.Description &&
		slices.EqualsPtr(s.TwoFactorDelete, o.TwoFactorDelete)
}

func (s *GenericClusterSpec) FromInstAPI(model *models.GenericClusterFields, version string) {
	s.Name = model.Name
	s.PrivateNetwork = model.PrivateNetworkCluster
	s.SLATier = model.SLATier
	s.Description = model.Description
	s.Version = version
	s.TwoFactorDeleteFromInstAPI(model.TwoFactorDelete)
}

func (s *GenericClusterSpec) TwoFactorDeleteFromInstAPI(instaModels []*models.TwoFactorDelete) {
	s.TwoFactorDelete = make([]*TwoFactorDelete, 0, len(instaModels))
	for _, instaModel := range instaModels {
		s.TwoFactorDelete = append(s.TwoFactorDelete, &TwoFactorDelete{
			Email: instaModel.ConfirmationEmail,
			Phone: instaModel.ConfirmationPhoneNumber,
		})
	}
}

func (s *GenericClusterSpec) ToInstAPI() models.GenericClusterFields {
	return models.GenericClusterFields{
		Name:                  s.Name,
		Description:           s.Description,
		PrivateNetworkCluster: s.PrivateNetwork,
		SLATier:               s.SLATier,
		TwoFactorDelete:       s.TwoFactorDeleteToInstAPI(),
	}
}

func (s *GenericClusterSpec) TwoFactorDeleteToInstAPI() []*models.TwoFactorDelete {
	tfd := make([]*models.TwoFactorDelete, 0, len(s.TwoFactorDelete))
	for _, t := range s.TwoFactorDelete {
		tfd = append(tfd, &models.TwoFactorDelete{
			ConfirmationPhoneNumber: t.Phone,
			ConfirmationEmail:       t.Email,
		})
	}

	return tfd
}

func (s *GenericClusterSpec) ClusterSettingsNeedUpdate(iCluster *GenericClusterSpec) bool {
	return len(s.TwoFactorDelete) != 0 && len(iCluster.TwoFactorDelete) == 0 ||
		s.Description != iCluster.Description
}

func (s *GenericClusterSpec) ClusterSettingsUpdateToInstAPI() *models.ClusterSettings {
	instaModels := &models.ClusterSettings{}
	if s.TwoFactorDelete != nil {
		instaModel := &models.TwoFactorDelete{}
		for _, tfd := range s.TwoFactorDelete {
			instaModel = tfd.ToInstAPI()
		}
		instaModels.TwoFactorDelete = instaModel
	}
	instaModels.Description = s.Description

	return instaModels
}

type GenericDataCentreSpec struct {
	// A logical name for the data centre within a cluster.
	// These names must be unique in the cluster.
	Name string `json:"name"`

	// Region of the Data Centre.
	Region string `json:"region"`

	// Name of a cloud provider service.
	CloudProvider string `json:"cloudProvider"`

	// For customers running in their own account.
	// Your provider account can be found on the Create Cluster page on the Instaclustr Console,
	// or the "Provider Account" property on any existing cluster.
	// For customers provisioning on Instaclustr's cloud provider accounts, this property may be omitted.
	//
	//+kubebuilder:default:=INSTACLUSTR
	ProviderAccountName string `json:"accountName,omitempty"`

	// The private network address block for the Data Centre specified using CIDR address notation.
	// The network must have a prefix length between /12 and /22 and must be part of a private address space.
	Network string `json:"network"`

	// List of tags to apply to the Data Centre.
	// Tags are metadata labels which allow you to identify, categorize and filter clusters.
	// This can be useful for grouping together clusters into applications, environments, or any category that you require.
	Tags map[string]string `json:"tags,omitempty"`

	// AWS specific settings for the Data Centre. Cannot be provided with GCP or Azure settings.
	//
	//+kubebuilder:validation:MaxItems:=1
	AWSSettings []*AWSSettings `json:"awsSettings,omitempty"`

	// GCP specific settings for the Data Centre. Cannot be provided with AWS or Azure settings.
	//
	//+kubebuilder:validation:MaxItems:=1
	GCPSettings []*GCPSettings `json:"gcpSettings,omitempty"`

	// Azure specific settings for the Data Centre. Cannot be provided with AWS or GCP settings.
	//
	//+kubebuilder:validation:MaxItems:=1
	AzureSettings []*AzureSettings `json:"azureSettings,omitempty"`
}

func (s *GenericDataCentreSpec) Equals(o *GenericDataCentreSpec) bool {
	return s.Name == o.Name &&
		s.Region == o.Region &&
		s.CloudProvider == o.CloudProvider &&
		s.ProviderAccountName == o.ProviderAccountName &&
		s.Network == o.Network &&
		areTagsEqual(s.Tags, o.Tags) &&
		slices.EqualsPtr(s.AWSSettings, o.AWSSettings) &&
		slices.EqualsPtr(s.GCPSettings, o.GCPSettings) &&
		slices.EqualsPtr(s.AzureSettings, o.AzureSettings)
}

func (s *GenericDataCentreSpec) FromInstAPI(model *models.GenericDataCentreFields) {
	s.Name = model.Name
	s.Region = model.Region
	s.CloudProvider = model.CloudProvider
	s.ProviderAccountName = model.ProviderAccountName
	s.Network = model.Network
	s.Tags = tagsFromInstAPI(model.Tags)
	s.cloudProviderSettingsFromInstAPI(model.CloudProviderSettings)
}

func (s *GenericDataCentreSpec) ToInstAPI() models.GenericDataCentreFields {
	return models.GenericDataCentreFields{
		Name:                  s.Name,
		Network:               s.Network,
		CloudProvider:         s.CloudProvider,
		Region:                s.Region,
		ProviderAccountName:   s.ProviderAccountName,
		Tags:                  tagsToInstAPI(s.Tags),
		CloudProviderSettings: s.cloudProviderSettingsToInstAPI(),
	}
}

func (s *GenericDataCentreSpec) cloudProviderSettingsToInstAPI() *models.CloudProviderSettings {
	var instaModel *models.CloudProviderSettings

	switch {
	case len(s.AWSSettings) > 0:
		setting := s.AWSSettings[0]
		instaModel = &models.CloudProviderSettings{AWSSettings: []*models.AWSSetting{{
			EBSEncryptionKey:       setting.DiskEncryptionKey,
			CustomVirtualNetworkID: setting.CustomVirtualNetworkID,
			BackupBucket:           setting.BackupBucket,
		}}}
	case len(s.GCPSettings) > 0:
		setting := s.GCPSettings[0]
		instaModel = &models.CloudProviderSettings{GCPSettings: []*models.GCPSetting{{
			CustomVirtualNetworkID:    setting.CustomVirtualNetworkID,
			DisableSnapshotAutoExpiry: setting.DisableSnapshotAutoExpiry,
		}}}
	case len(s.AzureSettings) > 0:
		setting := s.AzureSettings[0]
		instaModel = &models.CloudProviderSettings{AzureSettings: []*models.AzureSetting{{
			ResourceGroup:          setting.ResourceGroup,
			CustomVirtualNetworkID: setting.CustomVirtualNetworkID,
			StorageNetwork:         setting.StorageNetwork,
		}}}
	}

	return instaModel
}

func (s *GenericDataCentreSpec) cloudProviderSettingsFromInstAPI(instaModel *models.CloudProviderSettings) {
	if instaModel == nil {
		return
	}

	switch {
	case len(instaModel.AWSSettings) > 0 && instaModel.HasAWSCloudProviderSettings():
		setting := instaModel.AWSSettings[0]
		s.AWSSettings = []*AWSSettings{{
			DiskEncryptionKey:      setting.EBSEncryptionKey,
			CustomVirtualNetworkID: setting.CustomVirtualNetworkID,
			BackupBucket:           setting.BackupBucket,
		}}
	case len(instaModel.GCPSettings) > 0:
		setting := instaModel.GCPSettings[0]
		s.GCPSettings = []*GCPSettings{{
			CustomVirtualNetworkID:    setting.CustomVirtualNetworkID,
			DisableSnapshotAutoExpiry: setting.DisableSnapshotAutoExpiry,
		}}
	case len(instaModel.AzureSettings) > 0:
		setting := instaModel.AzureSettings[0]
		s.AzureSettings = []*AzureSettings{{
			ResourceGroup:          setting.ResourceGroup,
			CustomVirtualNetworkID: setting.CustomVirtualNetworkID,
			StorageNetwork:         setting.StorageNetwork,
		}}
	}
}
