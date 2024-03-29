/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// RedisUserV2 - Definition of a Redis User to be applied to a Redis cluster.
type RedisUserV2 struct {

	// Password for the Redis user.
	Password string `json:"password"`

	// ID of the Redis cluster.
	ClusterId string `json:"clusterId"`

	// Instaclustr identifier for the Redis user. The value of this property has the form: [cluster-id]_[redis-username]
	Id string `json:"id,omitempty"`

	// Permissions initially granted to Redis user upon creation.
	InitialPermissions string `json:"initialPermissions"`

	// Username of the Redis user.
	Username string `json:"username"`
}

// AssertRedisUserV2Required checks if the required fields are not zero-ed
func AssertRedisUserV2Required(obj RedisUserV2) error {
	elements := map[string]interface{}{
		"password":           obj.Password,
		"clusterId":          obj.ClusterId,
		"initialPermissions": obj.InitialPermissions,
		"username":           obj.Username,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertRedisUserV2Constraints checks if the values respects the defined constraints
func AssertRedisUserV2Constraints(obj RedisUserV2) error {
	return nil
}
