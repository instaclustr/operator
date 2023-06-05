/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// OpenSearchClusterV2 - Definition of a managed OpenSearch cluster that can be provisioned in Instaclustr.
type OpenSearchClusterV2 struct {

	// List of data node settings.
	DataNodes []OpenSearchDataNodeV2 `json:"dataNodes"`

	// Creates a PCI compliant cluster, see [PCI Compliance](https://www.instaclustr.com/support/documentation/useful-information/pci-compliance/).
	PciComplianceMode bool `json:"pciComplianceMode"`

	// Enable icu plugin
	IcuPlugin bool `json:"icuPlugin,omitempty"`

	// Version of OpenSearch to run on the cluster
	OpensearchVersion string `json:"opensearchVersion"`

	// Enable asynchronous search plugin
	AsynchronousSearchPlugin bool `json:"asynchronousSearchPlugin,omitempty"`

	//
	TwoFactorDelete []TwoFactorDeleteSettingsV2 `json:"twoFactorDelete,omitempty"`

	// Enable knn plugin
	KnnPlugin bool `json:"knnPlugin,omitempty"`

	// List of openSearch dashboards settings
	OpensearchDashboards []OpenSearchDashboardV2 `json:"opensearchDashboards,omitempty"`

	// Enable reporting plugin
	ReportingPlugin bool `json:"reportingPlugin,omitempty"`

	// Enable sql plugin
	SqlPlugin bool `json:"sqlPlugin,omitempty"`

	// Enable notifications plugin
	NotificationsPlugin bool `json:"notificationsPlugin,omitempty"`

	// List of data centre settings.
	DataCentres []OpenSearchDataCentreV2 `json:"dataCentres"`

	CurrentClusterOperationStatus CurrentClusterOperationStatusV2 `json:"currentClusterOperationStatus,omitempty"`

	// Enable anomaly detection plugin
	AnomalyDetectionPlugin bool `json:"anomalyDetectionPlugin,omitempty"`

	// Enable Load Balancer
	LoadBalancer bool `json:"loadBalancer,omitempty"`

	// Creates the cluster with private network only, see [Private Network Clusters](https://www.instaclustr.com/support/documentation/useful-information/private-network-clusters/).
	PrivateNetworkCluster bool `json:"privateNetworkCluster,omitempty"`

	// Name of the cluster.
	Name string `json:"name,omitempty"`

	// Provision this cluster for [Bundled Use only](https://www.instaclustr.com/support/documentation/cadence/getting-started-with-cadence/bundled-use-only-cluster-deployments/).
	BundledUseOnly bool `json:"bundledUseOnly,omitempty"`

	// List of cluster managers node settings
	ClusterManagerNodes []OpenSearchClusterManagerNodeV2 `json:"clusterManagerNodes,omitempty"`

	// Enable index management plugin
	IndexManagementPlugin bool `json:"indexManagementPlugin,omitempty"`

	// ID of the cluster.
	Id string `json:"id,omitempty"`

	SlaTier SlaTierV2 `json:"slaTier,omitempty"`

	// Enable alerting plugin
	AlertingPlugin bool `json:"alertingPlugin,omitempty"`

	// Status of the cluster.
	Status string `json:"status,omitempty"`
}

// AssertOpenSearchClusterV2Required checks if the required fields are not zero-ed
func AssertOpenSearchClusterV2Required(obj OpenSearchClusterV2) error {
	elements := map[string]interface{}{
		"pciComplianceMode": obj.PciComplianceMode,
		"opensearchVersion": obj.OpensearchVersion,
		"dataCentres":       obj.DataCentres,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	for _, el := range obj.DataNodes {
		if err := AssertOpenSearchDataNodeV2Required(el); err != nil {
			return err
		}
	}
	for _, el := range obj.TwoFactorDelete {
		if err := AssertTwoFactorDeleteSettingsV2Required(el); err != nil {
			return err
		}
	}
	for _, el := range obj.OpensearchDashboards {
		if err := AssertOpenSearchDashboardV2Required(el); err != nil {
			return err
		}
	}
	for _, el := range obj.DataCentres {
		if err := AssertOpenSearchDataCentreV2Required(el); err != nil {
			return err
		}
	}
	for _, el := range obj.ClusterManagerNodes {
		if err := AssertOpenSearchClusterManagerNodeV2Required(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertRecurseOpenSearchClusterV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of OpenSearchClusterV2 (e.g. [][]OpenSearchClusterV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseOpenSearchClusterV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aOpenSearchClusterV2, ok := obj.(OpenSearchClusterV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertOpenSearchClusterV2Required(aOpenSearchClusterV2)
	})
}
