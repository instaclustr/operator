/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type KafkaTopicDescribeConfigResponse struct {
	Topic string `json:"topic,omitempty"`

	Config map[string]string `json:"config,omitempty"`
}

// AssertKafkaTopicDescribeConfigResponseRequired checks if the required fields are not zero-ed
func AssertKafkaTopicDescribeConfigResponseRequired(obj KafkaTopicDescribeConfigResponse) error {
	return nil
}

// AssertKafkaTopicDescribeConfigResponseConstraints checks if the values respects the defined constraints
func AssertKafkaTopicDescribeConfigResponseConstraints(obj KafkaTopicDescribeConfigResponse) error {
	return nil
}
