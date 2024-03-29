/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type KafkaConsumerGroupState struct {
	ConsumerGroup string `json:"consumerGroup,omitempty"`

	ConsumerGroupState string `json:"consumerGroupState,omitempty"`

	ConsumerGroupClientDetails map[string][]string `json:"consumerGroupClientDetails,omitempty"`
}

// AssertKafkaConsumerGroupStateRequired checks if the required fields are not zero-ed
func AssertKafkaConsumerGroupStateRequired(obj KafkaConsumerGroupState) error {
	return nil
}

// AssertKafkaConsumerGroupStateConstraints checks if the values respects the defined constraints
func AssertKafkaConsumerGroupStateConstraints(obj KafkaConsumerGroupState) error {
	return nil
}
