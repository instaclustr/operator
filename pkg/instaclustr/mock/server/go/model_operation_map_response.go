/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type OperationMapResponse struct {
	Operations map[string]OperationResponse `json:"operations,omitempty"`
}

// AssertOperationMapResponseRequired checks if the required fields are not zero-ed
func AssertOperationMapResponseRequired(obj OperationMapResponse) error {
	return nil
}

// AssertOperationMapResponseConstraints checks if the values respects the defined constraints
func AssertOperationMapResponseConstraints(obj OperationMapResponse) error {
	return nil
}
