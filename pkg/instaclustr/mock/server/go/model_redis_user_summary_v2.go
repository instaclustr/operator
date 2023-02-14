/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
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

// AssertRecurseRedisUserSummaryV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of RedisUserSummaryV2 (e.g. [][]RedisUserSummaryV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseRedisUserSummaryV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aRedisUserSummaryV2, ok := obj.(RedisUserSummaryV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertRedisUserSummaryV2Required(aRedisUserSummaryV2)
	})
}
