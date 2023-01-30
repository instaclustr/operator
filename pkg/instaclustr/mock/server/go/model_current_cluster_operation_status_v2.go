/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// CurrentClusterOperationStatusV2 : Indicates if the cluster is currently performing any restructuring operation such as being created or resized
type CurrentClusterOperationStatusV2 string

// List of CurrentClusterOperationStatusV2
const (
	NO_OPERATION          CurrentClusterOperationStatusV2 = "NO_OPERATION"
	OPERATION_IN_PROGRESS CurrentClusterOperationStatusV2 = "OPERATION_IN_PROGRESS"
	OPERATION_FAILED      CurrentClusterOperationStatusV2 = "OPERATION_FAILED"
)

// AssertCurrentClusterOperationStatusV2Required checks if the required fields are not zero-ed
func AssertCurrentClusterOperationStatusV2Required(obj CurrentClusterOperationStatusV2) error {
	return nil
}

// AssertRecurseCurrentClusterOperationStatusV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of CurrentClusterOperationStatusV2 (e.g. [][]CurrentClusterOperationStatusV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseCurrentClusterOperationStatusV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aCurrentClusterOperationStatusV2, ok := obj.(CurrentClusterOperationStatusV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertCurrentClusterOperationStatusV2Required(aCurrentClusterOperationStatusV2)
	})
}
