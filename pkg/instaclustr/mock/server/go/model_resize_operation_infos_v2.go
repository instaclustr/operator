/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type ResizeOperationInfosV2 struct {
	Operations []ResizeOperationInfoV2 `json:"operations,omitempty"`
}

// AssertResizeOperationInfosV2Required checks if the required fields are not zero-ed
func AssertResizeOperationInfosV2Required(obj ResizeOperationInfosV2) error {
	for _, el := range obj.Operations {
		if err := AssertResizeOperationInfoV2Required(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertResizeOperationInfosV2Constraints checks if the values respects the defined constraints
func AssertResizeOperationInfosV2Constraints(obj ResizeOperationInfosV2) error {
	return nil
}
