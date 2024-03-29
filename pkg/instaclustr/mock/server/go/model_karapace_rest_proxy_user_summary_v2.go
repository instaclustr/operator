/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// KarapaceRestProxyUserSummaryV2 - Summary of a Karapace Rest Proxy User to be applied to a Kafka cluster.
type KarapaceRestProxyUserSummaryV2 struct {

	// ID of the Kafka cluster.
	ClusterId string `json:"clusterId"`

	// Instaclustr identifier for the Karapace Rest Proxy user. The value of this property has the form: [cluster-id]_[karapace-rest-proxy-username]
	Id string `json:"id,omitempty"`

	// Username of the Karapace Rest Proxy user.
	Username string `json:"username"`
}

// AssertKarapaceRestProxyUserSummaryV2Required checks if the required fields are not zero-ed
func AssertKarapaceRestProxyUserSummaryV2Required(obj KarapaceRestProxyUserSummaryV2) error {
	elements := map[string]interface{}{
		"clusterId": obj.ClusterId,
		"username":  obj.Username,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertKarapaceRestProxyUserSummaryV2Constraints checks if the values respects the defined constraints
func AssertKarapaceRestProxyUserSummaryV2Constraints(obj KarapaceRestProxyUserSummaryV2) error {
	return nil
}
