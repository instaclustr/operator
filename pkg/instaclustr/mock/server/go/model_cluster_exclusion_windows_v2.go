/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// ClusterExclusionWindowsV2 - Definition to the Cluster exclusion windows
type ClusterExclusionWindowsV2 struct {

	// List of cluster exclusion windows
	ExclusionWindows []ClusterExclusionWindowV2 `json:"exclusionWindows,omitempty"`

	// ID of the cluster
	ClusterId string `json:"clusterId,omitempty"`
}

// AssertClusterExclusionWindowsV2Required checks if the required fields are not zero-ed
func AssertClusterExclusionWindowsV2Required(obj ClusterExclusionWindowsV2) error {
	for _, el := range obj.ExclusionWindows {
		if err := AssertClusterExclusionWindowV2Required(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertClusterExclusionWindowsV2Constraints checks if the values respects the defined constraints
func AssertClusterExclusionWindowsV2Constraints(obj ClusterExclusionWindowsV2) error {
	return nil
}
