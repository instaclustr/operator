/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// OpenSearchPrivateLinkV2 -
type OpenSearchPrivateLinkV2 struct {
	IamPrincipalArns []string `json:"iamPrincipalArns,omitempty"`
}

// AssertOpenSearchPrivateLinkV2Required checks if the required fields are not zero-ed
func AssertOpenSearchPrivateLinkV2Required(obj OpenSearchPrivateLinkV2) error {
	return nil
}

// AssertRecurseOpenSearchPrivateLinkV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of OpenSearchPrivateLinkV2 (e.g. [][]OpenSearchPrivateLinkV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseOpenSearchPrivateLinkV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aOpenSearchPrivateLinkV2, ok := obj.(OpenSearchPrivateLinkV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertOpenSearchPrivateLinkV2Required(aOpenSearchPrivateLinkV2)
	})
}
