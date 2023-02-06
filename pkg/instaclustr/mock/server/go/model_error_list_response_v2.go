/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// ErrorListResponseV2 -
type ErrorListResponseV2 struct {
	Errors []ErrorResponseV2 `json:"errors,omitempty"`
}

// AssertErrorListResponseV2Required checks if the required fields are not zero-ed
func AssertErrorListResponseV2Required(obj ErrorListResponseV2) error {
	for _, el := range obj.Errors {
		if err := AssertErrorResponseV2Required(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertRecurseErrorListResponseV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of ErrorListResponseV2 (e.g. [][]ErrorListResponseV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseErrorListResponseV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aErrorListResponseV2, ok := obj.(ErrorListResponseV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertErrorListResponseV2Required(aErrorListResponseV2)
	})
}