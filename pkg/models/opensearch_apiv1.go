package models

const (
	OpenSearchV1                      = "OPENSEARCH"
	OpenSearchDataNodePurposeV1       = "OPENSEARCH_DATA_AND_INGEST"
	OpenSearchMasterNodePurposeV1     = "OPENSEARCH_MASTER"
	OpenSearchDashBoardsNodePurposeV1 = "OPENSEARCH_DASHBOARDS"
)

type OpenSearchClusterV1 struct {
	ID                    string                    `json:"id,omitempty"`
	ClusterName           string                    `json:"clusterName,omitempty"`
	ClusterStatus         string                    `json:"clusterStatus,omitempty"`
	ClusterProvider       []*ClusterProviderV1      `json:"clusterProvider,omitempty"`
	PrivateNetworkCluster bool                      `json:"privateNetworkCluster,omitempty"`
	SLATier               string                    `json:"slaTier,omitempty"`
	NodeSize              string                    `json:"nodeSize,omitempty"`
	ClusterNetwork        string                    `json:"clusterNetwork,omitempty"`
	DataCentre            string                    `json:"dataCentre,omitempty"`
	CDCID                 string                    `json:"cdcid,omitempty"`
	DataCentreCustomName  string                    `json:"dataCentreCustomName,omitempty"`
	FirewallRules         []*FirewallRule           `json:"firewallRules,omitempty"`
	OIDCProvider          string                    `json:"oidcProvider,omitempty"`
	BundledUseOnlyCluster bool                      `json:"bundledUseOnlyCluster,omitempty"`
	DataCentres           []*OpenSearchDataCentreV1 `json:"dataCentres,omitempty"`
	BundleOptions         *BundleOptions            `json:"bundleOptions,omitempty"`
	BundleVersion         string                    `json:"bundleVersion,omitempty"`
	PCICompliance         string                    `json:"pciCompliance,omitempty"`
	AddonBundles          []*AddonBundle            `json:"addonBundles,omitempty"`
	Bundles               []*OpenSearchBundleV1     `json:"bundles,omitempty"`
	Provider              *ClusterProviderV1        `json:"provider,omitempty"`
	PCICompliantCluster   bool                      `json:"pciCompliantCluster,omitempty"`
}

type OpenSearchCreateAPIv1 struct {
	ClusterName           string                `json:"clusterName,omitempty"`
	PrivateNetworkCluster bool                  `json:"privateNetworkCluster,omitempty"`
	SLATier               string                `json:"slaTier,omitempty"`
	NodeSize              string                `json:"nodeSize,omitempty"`
	ClusterNetwork        string                `json:"clusterNetwork,omitempty"`
	DataCentre            string                `json:"dataCentre,omitempty"`
	DataCentreCustomName  string                `json:"dataCentreCustomName,omitempty"`
	RackAllocation        *RackAllocationV1     `json:"rackAllocation,omitempty"`
	BundledUseOnlyCluster bool                  `json:"bundledUseOnlyCluster,omitempty"`
	Bundles               []*OpenSearchBundleV1 `json:"bundles,omitempty"`
	TwoFactorDelete       *TwoFactorDeleteV1    `json:"twoFactorDelete,omitempty"`
	Provider              *ClusterProviderV1    `json:"provider,omitempty"`
	PCICompliantCluster   bool                  `json:"pciCompliantCluster,omitempty"`
}

type OpenSearchBundleV1 struct {
	Bundle  `json:",inline"`
	Options *OpenSearchBundleOptionsV1 `json:"options"`
}

type OpenSearchDataCentreV1 struct {
	DataCentreV1 `json:",inline"`
}

type OpenSearchBundleOptionsV1 struct {
	DedicatedMasterNodes         bool   `json:"dedicatedMasterNodes,omitempty"`
	MasterNodeSize               string `json:"masterNodeSize,omitempty"`
	OpenSearchDashboardsNodeSize string `json:"openSearchDashboardsNodeSize,omitempty"`
	IndexManagementPlugin        bool   `json:"indexManagementPlugin,omitempty"`
	AlertingPlugin               bool   `json:"alertingPlugin,omitempty"`
	ICUPlugin                    bool   `json:"icuPlugin,omitempty"`
	KNNPlugin                    bool   `json:"knnPlugin,omitempty"`
	NotificationsPlugin          bool   `json:"notificationsPlugin,omitempty"`
	ReportsPlugin                bool   `json:"reportsPlugin,omitempty"`
}
