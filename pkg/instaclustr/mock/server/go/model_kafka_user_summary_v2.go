/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// KafkaUserSummaryV2 - Summary of a Kafka User to be applied to a Kafka cluster.
type KafkaUserSummaryV2 struct {

	// ID of the Kafka cluster.
	ClusterId string `json:"clusterId"`

	// Instaclustr identifier for the Kafka user. The value of this property has the form: [cluster-id]_[kafka-username]
	Id string `json:"id,omitempty"`

	// Username of the Kafka user.
	Username string `json:"username"`
}

// AssertKafkaUserSummaryV2Required checks if the required fields are not zero-ed
func AssertKafkaUserSummaryV2Required(obj KafkaUserSummaryV2) error {
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

// AssertRecurseKafkaUserSummaryV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of KafkaUserSummaryV2 (e.g. [][]KafkaUserSummaryV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseKafkaUserSummaryV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aKafkaUserSummaryV2, ok := obj.(KafkaUserSummaryV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertKafkaUserSummaryV2Required(aKafkaUserSummaryV2)
	})
}
