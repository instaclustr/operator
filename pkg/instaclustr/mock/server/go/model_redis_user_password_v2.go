/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
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

// AssertRedisUserPasswordV2Constraints checks if the values respects the defined constraints
func AssertRedisUserPasswordV2Constraints(obj RedisUserPasswordV2) error {
	return nil
}
