/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// OpenSearchDataNodeV2 -
type OpenSearchDataNodeV2 struct {

	// Size of data node
	NodeSize string `json:"nodeSize,omitempty"`

	// Number of nodes
	NodeCount int32 `json:"nodeCount,omitempty"`
}

// AssertOpenSearchDataNodeV2Required checks if the required fields are not zero-ed
func AssertOpenSearchDataNodeV2Required(obj OpenSearchDataNodeV2) error {
	return nil
}

// AssertRecurseOpenSearchDataNodeV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of OpenSearchDataNodeV2 (e.g. [][]OpenSearchDataNodeV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseOpenSearchDataNodeV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aOpenSearchDataNodeV2, ok := obj.(OpenSearchDataNodeV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertOpenSearchDataNodeV2Required(aOpenSearchDataNodeV2)
	})
}