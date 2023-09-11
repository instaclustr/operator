/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type OperationListResponse struct {
	Operations []OperationResponse `json:"operations,omitempty"`
}

// AssertOperationListResponseRequired checks if the required fields are not zero-ed
func AssertOperationListResponseRequired(obj OperationListResponse) error {
	for _, el := range obj.Operations {
		if err := AssertOperationResponseRequired(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertOperationListResponseConstraints checks if the values respects the defined constraints
func AssertOperationListResponseConstraints(obj OperationListResponse) error {
	return nil
}
