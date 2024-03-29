/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type KafkaTopicListResponse struct {
	Topics []string `json:"topics,omitempty"`
}

// AssertKafkaTopicListResponseRequired checks if the required fields are not zero-ed
func AssertKafkaTopicListResponseRequired(obj KafkaTopicListResponse) error {
	return nil
}

// AssertKafkaTopicListResponseConstraints checks if the values respects the defined constraints
func AssertKafkaTopicListResponseConstraints(obj KafkaTopicListResponse) error {
	return nil
}
