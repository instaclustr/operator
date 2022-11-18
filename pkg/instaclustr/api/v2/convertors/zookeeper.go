package convertors

import (
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
)

type CreateZookeeper struct {
	TwoFactorDelete       []*models.TwoFactorDelete `json:"twoFactorDelete,omitempty"`
	ZookeeperDataCentre   []*ZookeeperDataCentre    `json:"dataCentres"`
	PrivateNetworkCluster bool                      `json:"privateNetworkCluster"`
	ZookeeperVersion      string                    `json:"zookeeperVersion"`
	Name                  string                    `json:"name"`
	SLATier               string                    `json:"slaTier"`
}
type ZookeeperDataCentre struct {
	Name                     string                 `json:"name"`
	Network                  string                 `json:"network"`
	NodeSize                 string                 `json:"nodeSize"`
	NumberOfNodes            int32                  `json:"numberOfNodes"`
	AWSSettings              []*models.AWSSetting   `json:"awsSettings,omitempty"`
	GCPSettings              []*models.GCPSetting   `json:"gcpSettings,omitempty"`
	AzureSettings            []*models.AzureSetting `json:"azureSettings,omitempty"`
	Tags                     []*models.Tag          `json:"tags,omitempty"`
	CloudProvider            string                 `json:"cloudProvider"`
	Region                   string                 `json:"region"`
	ProviderAccountName      string                 `json:"providerAccountName,omitempty"`
	ClientToServerEncryption bool                   `json:"clientToServerEncryption"`
}

func ZookeeperToInstAPI(zook v1alpha1.ZookeeperSpec) CreateZookeeper {
	return CreateZookeeper{
		TwoFactorDelete:       twoFactorDeleteToInstAPI(zook.TwoFactorDelete),
		PrivateNetworkCluster: zook.PrivateNetworkCluster,
		Name:                  zook.Name,
		SLATier:               zook.SLATier,
		ZookeeperDataCentre:   zookeeperDataCentresToInstAPI(zook.DataCentres),
		ZookeeperVersion:      zook.Version,
	}
}

func zookeeperDataCentresToInstAPI(crdDCs []*v1alpha1.ZookeeperDataCentre) []*ZookeeperDataCentre {
	if crdDCs == nil {
		return nil
	}

	var instaDCs []*ZookeeperDataCentre

	for _, crdDC := range crdDCs {
		instaDC := &ZookeeperDataCentre{
			Name:                     crdDC.Name,
			Network:                  crdDC.Network,
			NodeSize:                 crdDC.NodeSize,
			NumberOfNodes:            crdDC.NodesNumber,
			Tags:                     TagsToInstAPI(crdDC.Tags),
			CloudProvider:            crdDC.CloudProvider,
			Region:                   crdDC.Region,
			ProviderAccountName:      crdDC.ProviderAccountName,
			ClientToServerEncryption: crdDC.ClientToServerEncryption,
		}

		allocateZookeeperProviderSettingsToInstAPI(crdDC, instaDC)

		instaDCs = append(instaDCs, instaDC)
	}

	return instaDCs
}

func allocateZookeeperProviderSettingsToInstAPI(crdDC *v1alpha1.ZookeeperDataCentre, instaDC *ZookeeperDataCentre) {
	for _, crdSetting := range crdDC.CloudProviderSettings {
		switch crdDC.CloudProvider {
		case models.AWSVPC:
			instaDC.AWSSettings = append(instaDC.AWSSettings, &models.AWSSetting{
				EBSEncryptionKey:       crdSetting.DiskEncryptionKey,
				CustomVirtualNetworkID: crdSetting.CustomVirtualNetworkID,
			})
		case models.GCP:
			instaDC.GCPSettings = append(instaDC.GCPSettings, &models.GCPSetting{
				CustomVirtualNetworkID: crdSetting.CustomVirtualNetworkID,
			})
		case models.AZURE, models.AZUREAZ:
			instaDC.AzureSettings = append(instaDC.AzureSettings, &models.AzureSetting{
				ResourceGroup: crdSetting.ResourceGroup,
			})
		}
	}
}
