package v1beta1

import (
	"github.com/instaclustr/operator/pkg/models"
	"github.com/instaclustr/operator/pkg/utils/slices"
)

type GenericClusterSpec struct {
	// Name [ 3 .. 32 ] characters.
	Name string `json:"name,omitempty"`

	Version string `json:"version,omitempty"`

	// The PCI compliance standards relate to the security of user data and transactional information.
	// Can only be applied clusters provisioned on AWS_VPC, running Cassandra, Kafka, Elasticsearch and Redis.
	PCICompliance bool `json:"pciCompliance,omitempty"`

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
		s.PCICompliance == o.PCICompliance &&
		s.PrivateNetwork == o.PrivateNetwork &&
		s.SLATier == o.SLATier &&
		s.Description == o.Description &&
		slices.EqualsPtr(s.TwoFactorDelete, o.TwoFactorDelete)
}

func (s *GenericClusterSpec) FromInstAPI(model *models.GenericClusterFields) {
	s.Name = model.Name
	s.PCICompliance = model.PCIComplianceMode
	s.PrivateNetwork = model.PrivateNetworkCluster
	s.SLATier = model.SLATier
	s.Description = model.Description
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
		PCIComplianceMode:     s.PCICompliance,
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
	Name                  string                   `json:"name,omitempty"`
	Region                string                   `json:"region"`
	CloudProvider         string                   `json:"cloudProvider"`
	ProviderAccountName   string                   `json:"accountName,omitempty"`
	Network               string                   `json:"network"`
	Tags                  map[string]string        `json:"tags,omitempty"`
	CloudProviderSettings []*CloudProviderSettings `json:"cloudProviderSettings,omitempty"`
}

func (s *GenericDataCentreSpec) Equals(o *GenericDataCentreSpec) bool {
	return s.Name == o.Name &&
		s.Region == o.Region &&
		s.CloudProvider == o.CloudProvider &&
		s.ProviderAccountName == o.ProviderAccountName &&
		s.Network == o.Network &&
		areTagsEqual(s.Tags, o.Tags) &&
		slices.EqualsPtr(s.CloudProviderSettings, o.CloudProviderSettings)
}

func (s *GenericDataCentreSpec) FromInstAPI(model *models.GenericDataCentreFields) {
	s.Name = model.Name
	s.Region = model.Region
	s.CloudProvider = model.CloudProvider
	s.ProviderAccountName = model.ProviderAccountName
	s.Network = model.Network
	s.Tags = tagsFromInstAPI(model.Tags)
	s.CloudProviderSettings = cloudProviderSettingsFromInstAPI(model)
}

func (dc *GenericDataCentreSpec) CloudProviderSettingsToInstAPI() models.CloudProviderSettings {
	instaModel := models.CloudProviderSettings{}

	switch dc.CloudProvider {
	case models.AWSVPC:
		for _, providerSettings := range dc.CloudProviderSettings {
			instaModel.AWSSettings = append(instaModel.AWSSettings, providerSettings.AWSToInstAPI())
		}
	case models.AZUREAZ:
		for _, providerSettings := range dc.CloudProviderSettings {
			instaModel.AzureSettings = append(instaModel.AzureSettings, providerSettings.AzureToInstAPI())
		}
	case models.GCP:
		for _, providerSettings := range dc.CloudProviderSettings {
			instaModel.GCPSettings = append(instaModel.GCPSettings, providerSettings.GCPToInstAPI())
		}
	}

	return instaModel
}

func (s *GenericDataCentreSpec) ToInstAPI() models.GenericDataCentreFields {
	return models.GenericDataCentreFields{
		Name:                  s.Name,
		Network:               s.Network,
		CloudProvider:         s.CloudProvider,
		Region:                s.Region,
		ProviderAccountName:   s.ProviderAccountName,
		Tags:                  tagsToInstAPI(s.Tags),
		CloudProviderSettings: s.CloudProviderSettingsToInstAPI(),
	}
}
