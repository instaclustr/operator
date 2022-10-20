package models

const (
	OpenSearch          = "OPENSEARCH"
	OpenSearchFinalizer = "opensearch-insta-finalizer/finalizer"
)

type OpenSearchBundle struct {
	Bundle  `json:",inline"`
	Options *OpenSearchBundleOptions `json:"options"`
}

type OpenSearchBundleOptions struct {
	DedicatedMasterNodes         bool   `json:"dedicatedMasterNodes,omitempty"`
	MasterNodeSize               string `json:"masterNodeSize,omitempty"`
	OpenSearchDashboardsNodeSize string `json:"openSearchDashboardsNodeSize,omitempty"`
	IndexManagementPlugin        bool   `json:"indexManagementPlugin,omitempty"`
}

type OpenSearchDataCentre struct {
	DataCentre `json:",inline"`
	Bundles    []*OpenSearchBundle `json:"bundles"`
}

type OpenSearchCluster struct {
	Cluster     `json:",inline"`
	Bundles     []*OpenSearchBundle     `json:"bundles"`
	DataCentres []*OpenSearchDataCentre `json:"dataCentres,omitempty"`
}

type OpenSearchStatus struct {
	ClusterStatus `json:",inline"`
	DataCentres   []*DataCentreStatus `json:"dataCentres,omitempty"`
}
