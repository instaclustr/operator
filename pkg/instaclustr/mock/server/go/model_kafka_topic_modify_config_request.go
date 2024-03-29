/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type KafkaTopicModifyConfigRequest struct {
	Config map[string]string `json:"config,omitempty"`

	ValidationMessages map[string]string `json:"validationMessages,omitempty"`
}

// AssertKafkaTopicModifyConfigRequestRequired checks if the required fields are not zero-ed
func AssertKafkaTopicModifyConfigRequestRequired(obj KafkaTopicModifyConfigRequest) error {
	return nil
}

// AssertKafkaTopicModifyConfigRequestConstraints checks if the values respects the defined constraints
func AssertKafkaTopicModifyConfigRequestConstraints(obj KafkaTopicModifyConfigRequest) error {
	return nil
}
