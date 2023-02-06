/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// ClusterSummaryV2 -
type ClusterSummaryV2 struct {

	//
	Application string `json:"application,omitempty"`

	//
	Id string `json:"id,omitempty"`
}

// AssertClusterSummaryV2Required checks if the required fields are not zero-ed
func AssertClusterSummaryV2Required(obj ClusterSummaryV2) error {
	return nil
}

// AssertRecurseClusterSummaryV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of ClusterSummaryV2 (e.g. [][]ClusterSummaryV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseClusterSummaryV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aClusterSummaryV2, ok := obj.(ClusterSummaryV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertClusterSummaryV2Required(aClusterSummaryV2)
	})
}