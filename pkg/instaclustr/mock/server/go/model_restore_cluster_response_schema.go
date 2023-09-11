/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type RestoreClusterResponseSchema struct {

	// Cluster ID
	RestoredCluster string `json:"restoredCluster,omitempty"`
}

// AssertRestoreClusterResponseSchemaRequired checks if the required fields are not zero-ed
func AssertRestoreClusterResponseSchemaRequired(obj RestoreClusterResponseSchema) error {
	return nil
}

// AssertRestoreClusterResponseSchemaConstraints checks if the values respects the defined constraints
func AssertRestoreClusterResponseSchemaConstraints(obj RestoreClusterResponseSchema) error {
	return nil
}
