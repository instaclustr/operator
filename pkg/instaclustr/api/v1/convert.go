package v1

import (
	"encoding/json"

	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	modelsv1 "github.com/instaclustr/operator/pkg/instaclustr/api/v1/models"
)

func ClusterStatusFromInstAPI(body []byte) (*v1alpha1.ClusterStatus, error) {
	var clusterStatusFromInst modelsv1.ClusterStatus
	err := json.Unmarshal(body, &clusterStatusFromInst)
	if err != nil {
		return nil, err
	}

	dataCentres := dataCentresFromInstAPI(clusterStatusFromInst.DataCentres)
	clusterStatus := &v1alpha1.ClusterStatus{
		ID:          clusterStatusFromInst.ID,
		Status:      clusterStatusFromInst.ClusterStatus,
		DataCentres: dataCentres,
	}

	return clusterStatus, nil
}

func PgToInstAPI(pgClusterSpec *v1alpha1.PgSpec) *modelsv1.PgCluster {
	dataCentresNumber := len(pgClusterSpec.DataCentres)
	isSingleDC := checkSingleDCCluster(dataCentresNumber)

	pgInstFirewallRules := firewallRulesToInstAPI(pgClusterSpec.FirewallRules)

	pgBundles := pgBundlesToInstAPI(pgClusterSpec.DataCentres[0], pgClusterSpec.Version, pgClusterSpec.PGBouncerVersion)

	pgInstProvider := pgProviderToInstAPI(pgClusterSpec.DataCentres[0])

	pgInstTwoFactorDelete := pgTwoFactorDeleteToInstAPI(pgClusterSpec.TwoFactorDelete)

	pg := &modelsv1.PgCluster{
		Cluster: modelsv1.Cluster{
			ClusterName:           pgClusterSpec.Name,
			NodeSize:              pgClusterSpec.DataCentres[0].NodeSize,
			PrivateNetworkCluster: pgClusterSpec.PrivateNetworkCluster,
			SLATier:               pgClusterSpec.SLATier,
			Provider:              pgInstProvider,
			FirewallRules:         pgInstFirewallRules,
			TwoFactorDelete:       pgInstTwoFactorDelete,
			OIDCProvider:          pgClusterSpec.OIDCProvider,
			BundledUseOnlyCluster: pgClusterSpec.BundledUseOnlyCluster,
		},
		Bundles: pgBundles,
	}

	if isSingleDC {
		pgRackAllocation := &modelsv1.RackAllocation{
			NodesPerRack:  pgClusterSpec.DataCentres[0].NodesNumber,
			NumberOfRacks: pgClusterSpec.DataCentres[0].RacksNumber,
		}

		pg.DataCentre = pgClusterSpec.DataCentres[0].Region
		pg.DataCentreCustomName = pgClusterSpec.DataCentres[0].Name
		pg.NodeSize = pgClusterSpec.DataCentres[0].NodeSize
		pg.ClusterNetwork = pgClusterSpec.DataCentres[0].Network
		pg.RackAllocation = pgRackAllocation

		return pg
	}

	var pgInstDCs []*modelsv1.PgDataCentre
	for _, dataCentre := range pgClusterSpec.DataCentres {
		pgBundles = pgBundlesToInstAPI(dataCentre, pgClusterSpec.Version, pgClusterSpec.PGBouncerVersion)

		pgInstProvider = pgProviderToInstAPI(dataCentre)

		pgRackAlloc := &modelsv1.RackAllocation{
			NodesPerRack:  dataCentre.NodesNumber,
			NumberOfRacks: dataCentre.RacksNumber,
		}

		pgInstDC := &modelsv1.PgDataCentre{
			DataCentre: modelsv1.DataCentre{
				Name:           dataCentre.Name,
				DataCentre:     dataCentre.Region,
				Network:        dataCentre.Network,
				Provider:       pgInstProvider,
				NodeSize:       dataCentre.NodeSize,
				RackAllocation: pgRackAlloc,
			},
			Bundles: pgBundles,
		}

		pgInstDCs = append(pgInstDCs, pgInstDC)
	}

	pg.DataCentres = pgInstDCs

	return pg
}

func pgBundlesToInstAPI(dataCentre *v1alpha1.PgDataCentre, version, pgBouncerVersion string) []*modelsv1.PgBundle {
	var pgBundles []*modelsv1.PgBundle
	pgBundle := &modelsv1.PgBundle{
		Bundle: modelsv1.Bundle{
			Bundle:  modelsv1.PgSQL,
			Version: version,
		},
		Options: &modelsv1.PgBundleOptions{
			ClientEncryption:      dataCentre.ClientEncryption,
			PostgresqlNodeCount:   dataCentre.PostgresqlNodeCount,
			ReplicationMode:       dataCentre.ReplicationMode,
			SynchronousModeStrict: dataCentre.SynchronousModeStrict,
		},
	}
	pgBundles = append(pgBundles, pgBundle)

	if pgBouncerVersion != "" {
		pgBouncerBundle := &modelsv1.PgBundle{
			Bundle: modelsv1.Bundle{
				Bundle:  modelsv1.PgBouncer,
				Version: pgBouncerVersion,
			},
			Options: &modelsv1.PgBundleOptions{
				PoolMode: dataCentre.PoolMode,
			},
		}
		pgBundles = append(pgBundles, pgBouncerBundle)
	}
	return pgBundles
}

func pgProviderToInstAPI(dataCentre *v1alpha1.PgDataCentre) *modelsv1.ClusterProvider {
	return &modelsv1.ClusterProvider{
		Name:                   dataCentre.CloudProvider,
		AccountName:            dataCentre.ProviderAccountName,
		CustomVirtualNetworkId: dataCentre.CloudProviderSettings[0].CustomVirtualNetworkId,
		Tags:                   dataCentre.Tags,
		ResourceGroup:          dataCentre.CloudProviderSettings[0].ResourceGroup,
		DiskEncryptionKey:      dataCentre.CloudProviderSettings[0].DiskEncryptionKey,
	}
}

func dataCentresFromInstAPI(instaDataCentres []*modelsv1.DataCentreStatus) []*v1alpha1.DataCentreStatus {
	var dataCentres []*v1alpha1.DataCentreStatus
	for _, dataCentre := range instaDataCentres {
		nodes := nodesFromInstAPI(dataCentre.Nodes)
		dataCentres = append(dataCentres, &v1alpha1.DataCentreStatus{
			ID:         dataCentre.ID,
			Status:     dataCentre.CDCStatus,
			Nodes:      nodes,
			NodeNumber: dataCentre.NodeCount,
		})
	}

	return dataCentres
}

func nodesFromInstAPI(instaNodes []*modelsv1.NodeStatus) []*v1alpha1.Node {
	var nodes []*v1alpha1.Node
	for _, node := range instaNodes {
		nodes = append(nodes, &v1alpha1.Node{
			ID:             node.ID,
			Rack:           node.Rack,
			PublicAddress:  node.PublicAddress,
			PrivateAddress: node.PrivateAddress,
			Status:         node.NodeStatus,
			Size:           node.Size,
		})
	}
	return nodes
}

func firewallRulesToInstAPI(firewallRules []*v1alpha1.FirewallRule) []*modelsv1.FirewallRule {
	if len(firewallRules) < 1 {
		return nil
	}

	var instFirewallRules []*modelsv1.FirewallRule

	for _, firewallRule := range firewallRules {
		var instRules []*modelsv1.RuleType
		for _, rule := range firewallRule.Rules {
			instRule := &modelsv1.RuleType{
				Type: rule.Type,
			}
			instRules = append(instRules, instRule)
		}

		instFirewallRule := &modelsv1.FirewallRule{
			Network: firewallRule.Network,
			Rules:   instRules,
		}

		instFirewallRules = append(instFirewallRules, instFirewallRule)
	}

	return instFirewallRules
}

func pgTwoFactorDeleteToInstAPI(twoFactorDelete []*v1alpha1.TwoFactorDelete) *modelsv1.TwoFactorDelete {
	if len(twoFactorDelete) < 1 {
		return nil
	}

	return &modelsv1.TwoFactorDelete{
		DeleteVerifyEmail: twoFactorDelete[0].Email,
		DeleteVerifyPhone: twoFactorDelete[0].Phone,
	}
}

func checkSingleDCCluster(dataCentresNumber int) bool {
	if dataCentresNumber > 1 {
		return false
	}

	return true
}
