/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type ListClusterforApi struct {

	// Cluster ID
	Id string `json:"id,omitempty"`

	// Cluster name
	Name string `json:"name,omitempty"`

	// Cassandra Bundle version
	CassandraVersion string `json:"cassandraVersion,omitempty"`

	// Number of nodes in the cluster
	NodeCount int32 `json:"nodeCount,omitempty"`

	// Number of nodes in the cluster in the \"RUNNING\" state
	RunningNodeCount int32 `json:"runningNodeCount,omitempty"`

	// Current Status of the cluster
	DerivedStatus string `json:"derivedStatus,omitempty"`

	// SLA tier of the cluster
	SlaTier string `json:"slaTier,omitempty"`

	// Cluster PCI compliance
	PciCompliance string `json:"pciCompliance,omitempty"`
}

// AssertListClusterforApiRequired checks if the required fields are not zero-ed
func AssertListClusterforApiRequired(obj ListClusterforApi) error {
	return nil
}

// AssertListClusterforApiConstraints checks if the values respects the defined constraints
func AssertListClusterforApiConstraints(obj ListClusterforApi) error {
	return nil
}
