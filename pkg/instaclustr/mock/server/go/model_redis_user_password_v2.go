/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// RedisUserPasswordV2 - Definition of a Redis User Password Update to be applied to a Redis user.
type RedisUserPasswordV2 struct {

	// Password for the Redis user.
	Password string `json:"password"`
}

// AssertRedisUserPasswordV2Required checks if the required fields are not zero-ed
func AssertRedisUserPasswordV2Required(obj RedisUserPasswordV2) error {
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

// AssertRecurseRedisUserPasswordV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of RedisUserPasswordV2 (e.g. [][]RedisUserPasswordV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseRedisUserPasswordV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aRedisUserPasswordV2, ok := obj.(RedisUserPasswordV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertRedisUserPasswordV2Required(aRedisUserPasswordV2)
	})
}