package instaclustr

import (
	"github.com/instaclustr/operator/apis/clusters/v1alpha1"
	"github.com/instaclustr/operator/pkg/instaclustr/models"
)

func PgToInstAPI(pgClusterSpec *v1alpha1.PgSpec) *models.PgClusterAPIv1 {
	dataCentresNumber := len(pgClusterSpec.DataCentres)
	isSingleDC := checkSingleDCCluster(dataCentresNumber)

	pgInstFirewallRules := firewallRulesToInstAPI(pgClusterSpec.FirewallRules)

	pgBundles := pgBundlesToInstAPI(pgClusterSpec.DataCentres[0], pgClusterSpec.Version, pgClusterSpec.PGBouncerVersion)

	pgInstProvider := pgProviderToInstAPI(pgClusterSpec.DataCentres[0])

	pgInstTwoFactorDelete := pgTwoFactorDeleteToInstAPI(pgClusterSpec.TwoFactorDelete)

	pg := &models.PgClusterAPIv1{
		ClusterAPIv1: models.ClusterAPIv1{
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
		pgRackAllocation := &models.RackAllocationAPIv1{
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

	var pgInstDCs []*models.PgDataCentreAPIv1
	for _, dataCentre := range pgClusterSpec.DataCentres {
		pgBundles = pgBundlesToInstAPI(dataCentre, pgClusterSpec.Version, pgClusterSpec.PGBouncerVersion)

		pgInstProvider = pgProviderToInstAPI(dataCentre)

		pgRackAlloc := &models.RackAllocationAPIv1{
			NodesPerRack:  dataCentre.NodesNumber,
			NumberOfRacks: dataCentre.RacksNumber,
		}

		pgInstDC := &models.PgDataCentreAPIv1{
			DataCentreAPIv1: models.DataCentreAPIv1{
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

func PgFromInstAPI(pgInstaCluster *models.PgStatusAPIv1) *v1alpha1.PgStatus {
	dataCentres := dataCentresFromInstAPIv1(pgInstaCluster.DataCentres)

	return &v1alpha1.PgStatus{
		ClusterStatus: v1alpha1.ClusterStatus{
			Status:                     pgInstaCluster.ClusterStatus,
			ID:                         pgInstaCluster.ID,
			ClusterCertificateDownload: pgInstaCluster.ClusterCertificateDownload,
			DataCentres:                dataCentres,
		},
	}
}

func pgBundlesToInstAPI(dataCentre *v1alpha1.PgDataCentre, version, pgBouncerVersion string) []*models.PgBundleAPIv1 {
	var pgBundles []*models.PgBundleAPIv1
	pgBundle := &models.PgBundleAPIv1{
		BundleAPIv1: models.BundleAPIv1{
			Bundle:  models.PgSQL,
			Version: version,
		},
		Options: &models.PgBundleOptionsAPIv1{
			ClientEncryption:      dataCentre.ClientEncryption,
			PostgresqlNodeCount:   dataCentre.PostgresqlNodeCount,
			ReplicationMode:       dataCentre.ReplicationMode,
			SynchronousModeStrict: dataCentre.SynchronousModeStrict,
		},
	}
	pgBundles = append(pgBundles, pgBundle)

	if pgBouncerVersion != "" {
		pgBouncerBundle := &models.PgBundleAPIv1{
			BundleAPIv1: models.BundleAPIv1{
				Bundle:  models.PgBouncer,
				Version: pgBouncerVersion,
			},
			Options: &models.PgBundleOptionsAPIv1{
				PoolMode: dataCentre.PoolMode,
			},
		}
		pgBundles = append(pgBundles, pgBouncerBundle)
	}
	return pgBundles
}

func pgProviderToInstAPI(dataCentre *v1alpha1.PgDataCentre) *models.ClusterProviderAPIv1 {
	return &models.ClusterProviderAPIv1{
		Name:                   dataCentre.Provider.Name,
		AccountName:            dataCentre.Provider.AccountName,
		CustomVirtualNetworkId: dataCentre.Provider.CustomVirtualNetworkId,
		Tags:                   dataCentre.Tags,
		ResourceGroup:          dataCentre.Provider.ResourceGroup,
		DiskEncryptionKey:      dataCentre.Provider.DiskEncryptionKey,
	}
}

func dataCentresFromInstAPIv1(instaDataCentres []*models.DataCentreStatusAPIv1) []*v1alpha1.DataCentreStatus {
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

func nodesFromInstAPIv1(instaNodes []*models.NodeStatusAPIv1) []*v1alpha1.Node {
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

func firewallRulesToInstAPI(firewallRules []*v1alpha1.FirewallRule) []*models.FirewallRuleAPIv1 {
	if len(firewallRules) < 1 {
		return nil
	}

	var instFirewallRules []*models.FirewallRuleAPIv1

	for _, firewallRule := range firewallRules {
		var instRules []*models.RuleTypeAPIv1
		for _, rule := range firewallRule.Rules {
			instRule := &models.RuleTypeAPIv1{
				Type: rule.Type,
			}
			instRules = append(instRules, instRule)
		}

		instFirewallRule := &models.FirewallRuleAPIv1{
			Network: firewallRule.Network,
			Rules:   instRules,
		}

		instFirewallRules = append(instFirewallRules, instFirewallRule)
	}

	return instFirewallRules
}

func pgTwoFactorDeleteToInstAPI(twoFactorDelete []*v1alpha1.TwoFactorDelete) *models.TwoFactorDeleteAPIv1 {
	if len(twoFactorDelete) < 1 {
		return nil
	}

	return &models.TwoFactorDeleteAPIv1{
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
