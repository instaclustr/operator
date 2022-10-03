package APIv1

import (
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr/APIv1/models"
)

func PgToInstAPI(pgClusterSpec *v1alpha1.PgSpec) *models.PgCluster {
	dataCentresNumber := len(pgClusterSpec.DataCentres)
	isSingleDC := checkSingleDCCluster(dataCentresNumber)

	pgInstFirewallRules := firewallRulesToInstAPI(pgClusterSpec.FirewallRules)

	pgBundles := pgBundlesToInstAPI(pgClusterSpec.DataCentres[0], pgClusterSpec.Version, pgClusterSpec.PGBouncerVersion)

	pgInstProvider := pgProviderToInstAPI(pgClusterSpec.DataCentres[0])

	pgInstTwoFactorDelete := pgTwoFactorDeleteToInstAPI(pgClusterSpec.TwoFactorDelete)

	pg := &models.PgCluster{
		Cluster: models.Cluster{
			ClusterName:           pgClusterSpec.ClusterName,
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
		pgRackAllocation := &models.RackAllocation{
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

	var pgInstDCs []*models.PgDataCentre
	for _, dataCentre := range pgClusterSpec.DataCentres {
		pgBundles = pgBundlesToInstAPI(dataCentre, pgClusterSpec.Version, pgClusterSpec.PGBouncerVersion)

		pgInstProvider = pgProviderToInstAPI(dataCentre)

		pgRackAlloc := &models.RackAllocation{
			NodesPerRack:  dataCentre.NodesNumber,
			NumberOfRacks: dataCentre.RacksNumber,
		}

		pgInstDC := &models.PgDataCentre{
			DataCentre: models.DataCentre{
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

func PgFromInstAPI(pgInstaCluster *models.PgStatus) *v1alpha1.PgStatus {
	dataCentres := dataCentresFromInstAPIv1(pgInstaCluster.DataCentres)

	return &v1alpha1.PgStatus{
		ClusterStatus: v1alpha1.ClusterStatus{
			Status:                     pgInstaCluster.ClusterStatus.ClusterStatus,
			ID:                         pgInstaCluster.ID,
			ClusterCertificateDownload: pgInstaCluster.ClusterCertificateDownload,
			DataCentres:                dataCentres,
		},
	}
}

func pgBundlesToInstAPI(dataCentre *v1alpha1.PgDataCentre, version, pgBouncerVersion string) []*models.PgBundle {
	var pgBundles []*models.PgBundle
	pgBundle := &models.PgBundle{
		Bundle: models.Bundle{
			Bundle:  models.PgSQL,
			Version: version,
		},
		Options: &models.PgBundleOptions{
			ClientEncryption:      dataCentre.ClientEncryption,
			PostgresqlNodeCount:   dataCentre.PostgresqlNodeCount,
			ReplicationMode:       dataCentre.ReplicationMode,
			SynchronousModeStrict: dataCentre.SynchronousModeStrict,
		},
	}
	pgBundles = append(pgBundles, pgBundle)

	if pgBouncerVersion != "" {
		pgBouncerBundle := &models.PgBundle{
			Bundle: models.Bundle{
				Bundle:  models.PgBouncer,
				Version: pgBouncerVersion,
			},
			Options: &models.PgBundleOptions{
				PoolMode: dataCentre.PoolMode,
			},
		}
		pgBundles = append(pgBundles, pgBouncerBundle)
	}
	return pgBundles
}

func pgProviderToInstAPI(dataCentre *v1alpha1.PgDataCentre) *models.ClusterProvider {
	return &models.ClusterProvider{
		Name:                   dataCentre.Provider.Name,
		AccountName:            dataCentre.Provider.AccountName,
		CustomVirtualNetworkId: dataCentre.Provider.CustomVirtualNetworkId,
		Tags:                   dataCentre.Tags,
		ResourceGroup:          dataCentre.Provider.ResourceGroup,
		DiskEncryptionKey:      dataCentre.Provider.DiskEncryptionKey,
	}
}

func dataCentresFromInstAPIv1(instaDataCentres []*models.DataCentreStatus) []*v1alpha1.DataCentreStatus {
	var dataCentres []*v1alpha1.DataCentreStatus
	for _, dataCentre := range instaDataCentres {
		nodes := nodesFromInstAPIv1(dataCentre.Nodes)
		dataCentres = append(dataCentres, &v1alpha1.DataCentreStatus{
			ID:        dataCentre.ID,
			Status:    dataCentre.Status,
			Nodes:     nodes,
			NodeCount: dataCentre.NodeCount,
		})
	}

	return dataCentres
}

func nodesFromInstAPIv1(instaNodes []*models.NodeStatus) []*v1alpha1.Node {
	var nodes []*v1alpha1.Node
	for _, node := range instaNodes {
		nodes = append(nodes, &v1alpha1.Node{
			NodeID:         node.ID,
			Rack:           node.Rack,
			PublicAddress:  node.PublicAddress,
			PrivateAddress: node.PrivateAddress,
			NodeStatus:     node.Status,
			NodeSize:       node.Size,
		})
	}
	return nodes
}

func firewallRulesToInstAPI(firewallRules []*v1alpha1.FirewallRule) []*models.FirewallRule {
	if len(firewallRules) < 1 {
		return nil
	}

	var instFirewallRules []*models.FirewallRule

	for _, firewallRule := range firewallRules {
		var instRules []*models.RuleType
		for _, rule := range firewallRule.Rules {
			instRule := &models.RuleType{
				Type: rule.Type,
			}
			instRules = append(instRules, instRule)
		}

		instFirewallRule := &models.FirewallRule{
			Network: firewallRule.Network,
			Rules:   instRules,
		}

		instFirewallRules = append(instFirewallRules, instFirewallRule)
	}

	return instFirewallRules
}

func pgTwoFactorDeleteToInstAPI(twoFactorDelete []*v1alpha1.TwoFactorDelete) *models.TwoFactorDelete {
	if len(twoFactorDelete) < 1 {
		return nil
	}

	return &models.TwoFactorDelete{
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
