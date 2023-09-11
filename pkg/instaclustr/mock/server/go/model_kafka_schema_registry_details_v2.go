/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type KafkaSchemaRegistryDetailsV2 struct {

	// Adds the specified version of Kafka Schema Registry to the Kafka cluster. Available versions: <ul> <li>`5.0.0`</li> </ul>
	Version string `json:"version"`
}

// AssertKafkaSchemaRegistryDetailsV2Required checks if the required fields are not zero-ed
func AssertKafkaSchemaRegistryDetailsV2Required(obj KafkaSchemaRegistryDetailsV2) error {
	elements := map[string]interface{}{
		"version": obj.Version,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertKafkaSchemaRegistryDetailsV2Constraints checks if the values respects the defined constraints
func AssertKafkaSchemaRegistryDetailsV2Constraints(obj KafkaSchemaRegistryDetailsV2) error {
	return nil
}
