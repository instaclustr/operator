package convertors

import (
	"encoding/json"

	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
)

func twoFactorDeleteToInstAPI(crdTwoFactors []*v1alpha1.TwoFactorDelete) []*modelsv2.TwoFactorDelete {
	if len(crdTwoFactors) < 1 {
		return nil
	}

	var instaFactorDelete []*modelsv2.TwoFactorDelete
	for _, twoFactor := range crdTwoFactors {
		instaFactorDelete = append(instaFactorDelete, &modelsv2.TwoFactorDelete{
			ConfirmationPhoneNumber: twoFactor.Phone,
			ConfirmationEmail:       twoFactor.Email,
		})
	}

	return instaFactorDelete
}

func TagsToInstAPI(crdTags map[string]string) []*modelsv2.Tag {
	var instaTags []*modelsv2.Tag

	for k, v := range crdTags {
		instaTags = append(instaTags, &modelsv2.Tag{
			Key:   k,
			Value: v,
		})
	}

	return instaTags
}

func ClusterStatusFromInstAPI(body []byte) (*v1alpha1.FullClusterStatus, error) {
	var clusterStatusFromInst modelsv2.ClusterStatus
	err := json.Unmarshal(body, &clusterStatusFromInst)
	if err != nil {
		return nil, err
	}

	dataCentres := dataCentresFromInstAPI(clusterStatusFromInst.DataCentres)
	clusterStatus := &v1alpha1.FullClusterStatus{
		ClusterStatus: v1alpha1.ClusterStatus{
			ID:          clusterStatusFromInst.ID,
			Status:      clusterStatusFromInst.Status,
			DataCentres: dataCentres,
		},
	}

	return clusterStatus, nil
}

func dataCentresFromInstAPI(instaDataCentres []*modelsv2.DataCentreStatus) []*v1alpha1.DataCentreStatus {
	var dataCentres []*v1alpha1.DataCentreStatus
	for _, dataCentre := range instaDataCentres {
		nodes := nodesFromInstAPI(dataCentre.Nodes)
		dataCentres = append(dataCentres, &v1alpha1.DataCentreStatus{
			ID:         dataCentre.ID,
			Status:     dataCentre.Status,
			Nodes:      nodes,
			NodeNumber: dataCentre.NumberOfNodes,
		})
	}

	return dataCentres
}

func nodesFromInstAPI(instaNodes []*modelsv2.Node) []*v1alpha1.Node {
	var nodes []*v1alpha1.Node
	for _, node := range instaNodes {
		nodes = append(nodes, &v1alpha1.Node{
			ID:             node.ID,
			Rack:           node.Rack,
			PublicAddress:  node.PublicAddress,
			PrivateAddress: node.PrivateAddress,
			Status:         node.Status,
			Size:           node.Size,
		})
	}
	return nodes
}

func dataCentresToInstAPI(crdDCs []*v1alpha1.DataCentre) []*modelsv2.DataCentre {
	if crdDCs == nil {
		return nil
	}

	var instaDCs []*modelsv2.DataCentre

	for _, crdDC := range crdDCs {
		instaDC := &modelsv2.DataCentre{
			Name:                crdDC.Name,
			Network:             crdDC.Network,
			NodeSize:            crdDC.NodeSize,
			NumberOfNodes:       crdDC.NodesNumber,
			Tags:                TagsToInstAPI(crdDC.Tags),
			CloudProvider:       crdDC.CloudProvider,
			Region:              crdDC.Region,
			ProviderAccountName: crdDC.ProviderAccountName,
		}

		allocateProviderSettingsToInstAPI(crdDC, instaDC)

		instaDCs = append(instaDCs, instaDC)
	}

	return instaDCs
}

func allocateProviderSettingsToInstAPI(crdDC *v1alpha1.DataCentre, instaDC *modelsv2.DataCentre) {
	for _, crdSetting := range crdDC.CloudProviderSettings {
		switch crdDC.CloudProvider {
		case modelsv2.AWSVPC:
			instaDC.AWSSettings = append(instaDC.AWSSettings, &modelsv2.AWSSetting{
				EBSEncryptionKey:       crdSetting.DiskEncryptionKey,
				CustomVirtualNetworkID: crdSetting.CustomVirtualNetworkID,
			})
		case modelsv2.GCP:
			instaDC.GCPSettings = append(instaDC.GCPSettings, &modelsv2.GCPSetting{
				CustomVirtualNetworkID: crdSetting.CustomVirtualNetworkID,
			})
		case modelsv2.AZURE, modelsv2.AZUREAZ:
			instaDC.AzureSettings = append(instaDC.AzureSettings, &modelsv2.AzureSetting{
				ResourceGroup: crdSetting.ResourceGroup,
			})
		}
	}

}
