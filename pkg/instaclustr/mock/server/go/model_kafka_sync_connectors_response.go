/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type KafkaSyncConnectorsResponse struct {

	// Status of the Kafka sync connectors request
	Status string `json:"status,omitempty"`
}

// AssertKafkaSyncConnectorsResponseRequired checks if the required fields are not zero-ed
func AssertKafkaSyncConnectorsResponseRequired(obj KafkaSyncConnectorsResponse) error {
	return nil
}

// AssertKafkaSyncConnectorsResponseConstraints checks if the values respects the defined constraints
func AssertKafkaSyncConnectorsResponseConstraints(obj KafkaSyncConnectorsResponse) error {
	return nil
}