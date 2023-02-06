/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// KafkaUserOptionsV2 - Initial options used when creating Kafka user
type KafkaUserOptionsV2 struct {

	// Overwrite user if already exists.
	OverrideExistingUser bool `json:"overrideExistingUser,omitempty"`

	// SASL/SCRAM mechanism for user
	SaslScramMechanism string `json:"saslScramMechanism"`
}

// AssertKafkaUserOptionsV2Required checks if the required fields are not zero-ed
func AssertKafkaUserOptionsV2Required(obj KafkaUserOptionsV2) error {
	elements := map[string]interface{}{
		"saslScramMechanism": obj.SaslScramMechanism,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertRecurseKafkaUserOptionsV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of KafkaUserOptionsV2 (e.g. [][]KafkaUserOptionsV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseKafkaUserOptionsV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aKafkaUserOptionsV2, ok := obj.(KafkaUserOptionsV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertKafkaUserOptionsV2Required(aKafkaUserOptionsV2)
	})
}