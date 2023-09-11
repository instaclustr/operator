/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// KafkaListenerProtocolEnumV2 : Kafka listener protocols
type KafkaListenerProtocolEnumV2 string

// List of KafkaListenerProtocolEnumV2
const (
	SASL_SSL       KafkaListenerProtocolEnumV2 = "SASL_SSL"
	SASL_PLAINTEXT KafkaListenerProtocolEnumV2 = "SASL_PLAINTEXT"
	SSL            KafkaListenerProtocolEnumV2 = "SSL"
	PLAINTEXT      KafkaListenerProtocolEnumV2 = "PLAINTEXT"
)

// AssertKafkaListenerProtocolEnumV2Required checks if the required fields are not zero-ed
func AssertKafkaListenerProtocolEnumV2Required(obj KafkaListenerProtocolEnumV2) error {
	return nil
}

// AssertKafkaListenerProtocolEnumV2Constraints checks if the values respects the defined constraints
func AssertKafkaListenerProtocolEnumV2Constraints(obj KafkaListenerProtocolEnumV2) error {
	return nil
}
