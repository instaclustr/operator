package v2

import (
	"encoding/json"

	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	modelsv2 "github.com/instaclustr/operator/pkg/instaclustr/api/v2/models"
)

func CassandraToInstAPI(cassandraSpec *v1alpha1.CassandraSpec) *modelsv2.CassandraCluster {
	cassandraInstTwoFactorDelete := cassandraTwoFactorDeleteToInstAPI(cassandraSpec.TwoFactorDelete)

	cassandra := &modelsv2.CassandraCluster{
		ClusterSpec: modelsv2.ClusterSpec{
			Name:                  cassandraSpec.Name,
			SLATier:               cassandraSpec.SLATier,
			PrivateNetworkCluster: cassandraSpec.PrivateNetworkCluster,
			PCIComplianceMode:     cassandraSpec.PCICompliance,
			TwoFactorDeletes:      cassandraInstTwoFactorDelete,
		},
		CassandraVersion:    cassandraSpec.Version,
		LuceneEnabled:       cassandraSpec.LuceneEnabled,
		PasswordAndUserAuth: cassandraSpec.PasswordAndUserAuth,
	}

	var cassandraInstDCs []*modelsv2.CassandraDataCentre
	for _, dataCentre := range cassandraSpec.DataCentres {
		awsSettings, gcpSettings, azureSettings := providerSettingsToInstaCloudProviders(dataCentre)

		cassandraInstDC := &modelsv2.CassandraDataCentre{
			DataCentre: modelsv2.DataCentre{
				Name:                dataCentre.Name,
				Network:             dataCentre.Network,
				NodeSize:            dataCentre.NodeSize,
				NumberOfNodes:       dataCentre.NodesNumber,
				Region:              dataCentre.Region,
				Tags:                tagsToInstAPI(dataCentre.Tags),
				CloudProvider:       dataCentre.CloudProvider,
				AWSSettings:         awsSettings,
				GCPSettings:         gcpSettings,
				AzureSettings:       azureSettings,
				ProviderAccountName: dataCentre.ProviderAccountName,
			},
			ContinuousBackup:               dataCentre.ContinuousBackup,
			PrivateIPBroadcastForDiscovery: dataCentre.PrivateIPBroadcastForDiscovery,
			ClientToClusterEncryption:      dataCentre.ClientToClusterEncryption,
			ReplicationFactor:              dataCentre.RacksNumber,
		}

		cassandraInstDCs = append(cassandraInstDCs, cassandraInstDC)

	}
	cassandra.DataCentres = cassandraInstDCs

	return cassandra
}

func cassandraTwoFactorDeleteToInstAPI(twoFactorDelete []*v1alpha1.TwoFactorDelete) []*modelsv2.TwoFactorDelete {
	if len(twoFactorDelete) < 1 {
		return nil
	}

	var twoFactor []*modelsv2.TwoFactorDelete
	for i := range twoFactorDelete {
		twoFactor[i].ConfirmationEmail = twoFactorDelete[i].Email
		twoFactor[i].ConfirmationPhoneNumber = twoFactorDelete[i].Phone
	}

	return twoFactor
}

func tagsToInstAPI(tags map[string]string) []*modelsv2.Tag {
	var res []*modelsv2.Tag

	for k, v := range tags {
		res = append(res, &modelsv2.Tag{
			Key:   k,
			Value: v,
		})
	}

	return res
}

func providerSettingsToInstaCloudProviders(
	dataCentre *v1alpha1.CassandraDataCentre,
) ([]*modelsv2.AWSSetting, []*modelsv2.GCPSetting, []*modelsv2.AzureSetting) {
	var AWSSettings []*modelsv2.AWSSetting
	var GCPSettings []*modelsv2.GCPSetting
	var AzureSettings []*modelsv2.AzureSetting

	for _, cp := range dataCentre.CloudProviderSettings {
		if dataCentre.CloudProvider == "AWS_VPC" {
			AWSSetting := &modelsv2.AWSSetting{
				CustomVirtualNetworkID: cp.CustomVirtualNetworkId,
				EBSEncryptionKey:       cp.DiskEncryptionKey,
			}
			AWSSettings = append(AWSSettings, AWSSetting)
		} else if dataCentre.CloudProvider == "GCP" {
			GCPSetting := &modelsv2.GCPSetting{
				CustomVirtualNetworkID: cp.CustomVirtualNetworkId,
			}
			GCPSettings = append(GCPSettings, GCPSetting)
		} else if dataCentre.CloudProvider == "AZURE_AZ" {
			AzureSetting := &modelsv2.AzureSetting{
				ResourceGroup: cp.ResourceGroup,
			}
			AzureSettings = append(AzureSettings, AzureSetting)
		}

	}

	return AWSSettings, GCPSettings, AzureSettings
}

func ClusterStatusFromInstAPI(body []byte) (*v1alpha1.ClusterStatus, error) {
	var clusterStatusFromInst modelsv2.ClusterStatus
	err := json.Unmarshal(body, &clusterStatusFromInst)
	if err != nil {
		return nil, err
	}

	dataCentres := dataCentresFromInstAPI(clusterStatusFromInst.DataCentres)
	clusterStatus := &v1alpha1.ClusterStatus{
		ID:          clusterStatusFromInst.ID,
		Status:      clusterStatusFromInst.Status,
		DataCentres: dataCentres,
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
