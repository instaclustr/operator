/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// SlaTierV2 : SLA Tier of the cluster. Non-production clusters may receive lower priority support and reduced SLAs. Production tier is not available when using Developer class nodes. See [SLA Tier](https://www.instaclustr.com/support/documentation/useful-information/sla-tier/) for more information.
type SlaTierV2 string

// List of SlaTierV2
const (
	PRODUCTION     SlaTierV2 = "PRODUCTION"
	NON_PRODUCTION SlaTierV2 = "NON_PRODUCTION"
)

// AssertSlaTierV2Required checks if the required fields are not zero-ed
func AssertSlaTierV2Required(obj SlaTierV2) error {
	return nil
}

// AssertRecurseSlaTierV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of SlaTierV2 (e.g. [][]SlaTierV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseSlaTierV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aSlaTierV2, ok := obj.(SlaTierV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertSlaTierV2Required(aSlaTierV2)
	})
}