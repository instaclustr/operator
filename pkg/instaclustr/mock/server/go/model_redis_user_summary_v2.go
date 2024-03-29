/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// RedisUserSummaryV2 - Summary of a Redis User to be applied to a Redis cluster.
type RedisUserSummaryV2 struct {

	// ID of the Redis cluster.
	ClusterId string `json:"clusterId"`

	// Instaclustr identifier for the Redis user. The value of this property has the form: [cluster-id]_[redis-username]
	Id string `json:"id,omitempty"`

	// Username of the Redis user.
	Username string `json:"username"`
}

// AssertRedisUserSummaryV2Required checks if the required fields are not zero-ed
func AssertRedisUserSummaryV2Required(obj RedisUserSummaryV2) error {
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

// AssertRedisUserSummaryV2Constraints checks if the values respects the defined constraints
func AssertRedisUserSummaryV2Constraints(obj RedisUserSummaryV2) error {
	return nil
}
