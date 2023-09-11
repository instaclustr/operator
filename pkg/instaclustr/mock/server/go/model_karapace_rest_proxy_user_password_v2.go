/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// KarapaceRestProxyUserPasswordV2 - Definition of a Karapace Rest Proxy User Password Update to be applied to a Karapace Rest Proxy user.
type KarapaceRestProxyUserPasswordV2 struct {

	// Password for the Karapace Rest Proxy user.
	Password string `json:"password"`
}

// AssertKarapaceRestProxyUserPasswordV2Required checks if the required fields are not zero-ed
func AssertKarapaceRestProxyUserPasswordV2Required(obj KarapaceRestProxyUserPasswordV2) error {
	elements := map[string]interface{}{
		"password": obj.Password,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertKarapaceRestProxyUserPasswordV2Constraints checks if the values respects the defined constraints
func AssertKarapaceRestProxyUserPasswordV2Constraints(obj KarapaceRestProxyUserPasswordV2) error {
	return nil
}
