/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type KafkaConnectExternalSourceClusterDetailsV2 struct {

	// Kafka connection properties string used to connect to external kafka cluster
	SourceConnectionProperties string `json:"sourceConnectionProperties"`
}

// AssertKafkaConnectExternalSourceClusterDetailsV2Required checks if the required fields are not zero-ed
func AssertKafkaConnectExternalSourceClusterDetailsV2Required(obj KafkaConnectExternalSourceClusterDetailsV2) error {
	elements := map[string]interface{}{
		"sourceConnectionProperties": obj.SourceConnectionProperties,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertKafkaConnectExternalSourceClusterDetailsV2Constraints checks if the values respects the defined constraints
func AssertKafkaConnectExternalSourceClusterDetailsV2Constraints(obj KafkaConnectExternalSourceClusterDetailsV2) error {
	return nil
}
