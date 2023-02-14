package models

const (
	OpenSearchV1                      = "OPENSEARCH"
	OpenSearchDataNodePurposeV1       = "OPENSEARCH_DATA_AND_INGEST"
	OpenSearchMasterNodePurposeV1     = "OPENSEARCH_MASTER"
	OpenSearchDashBoardsNodePurposeV1 = "OPENSEARCH_DASHBOARDS"
)

type OpenSearchBundleV1 struct {
	Bundle  `json:",inline"`
	Options *OpenSearchBundleOptionsV1 `json:"options"`
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

type OpenSearchClusterCreationV1 struct {
	ClusterV1       `json:",inline"`
	Bundles         []*OpenSearchBundleV1 `json:"bundles"`
	TwoFactorDelete *TwoFactorDelete      `json:"twoFactorDelete,omitempty"`
	Provider        *ClusterProviderV1    `json:"provider,omitempty"`
}
