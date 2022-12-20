package models

import "github.com/instaclustr/operator/pkg/models"

const (
	OpenSearch                      = "OPENSEARCH"
	OpenSearchDataNodePurpose       = "OPENSEARCH_DATA_AND_INGEST"
	OpenSearchMasterNodePurpose     = "OPENSEARCH_MASTER"
	OpenSearchDashBoardsNodePurpose = "OPENSEARCH_DASHBOARDS"
)

type OpenSearchBundle struct {
	models.Bundle `json:",inline"`
	Options       *OpenSearchBundleOptions `json:"options"`
}

type OpenSearchBundleOptions struct {
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

type OpenSearchDataCentre struct {
	models.DataCentre `json:",inline"`
	Bundles           []*OpenSearchBundle `json:"bundles"`
}

type OpenSearchCluster struct {
	models.Cluster `json:",inline"`
	Bundles        []*OpenSearchBundle     `json:"bundles"`
	DataCentres    []*OpenSearchDataCentre `json:"dataCentres,omitempty"`
}

type OpenSearchStatus struct {
	models.ClusterStatus `json:",inline"`
	DataCentres          []*models.DataCentreStatus `json:"dataCentres,omitempty"`
}
