/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// KafkaListenerTypeEnumV2 : Kafka listener types
type KafkaListenerTypeEnumV2 string

// List of KafkaListenerTypeEnumV2
const (
	PUBLIC  KafkaListenerTypeEnumV2 = "PUBLIC"
	PRIVATE KafkaListenerTypeEnumV2 = "PRIVATE"
	MTLS    KafkaListenerTypeEnumV2 = "MTLS"
)

// AssertKafkaListenerTypeEnumV2Required checks if the required fields are not zero-ed
func AssertKafkaListenerTypeEnumV2Required(obj KafkaListenerTypeEnumV2) error {
	return nil
}

// AssertKafkaListenerTypeEnumV2Constraints checks if the values respects the defined constraints
func AssertKafkaListenerTypeEnumV2Constraints(obj KafkaListenerTypeEnumV2) error {
	return nil
}
