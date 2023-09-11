/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type KafkaRestProxyOptionsSchema struct {

	// Accepts true/false. Enables Integration of the REST proxy with a Schema registry
	IntegrateRestProxyWithSchemaRegistry bool `json:"integrateRestProxyWithSchemaRegistry,omitempty"`

	// Accepts true/false. Integrates the REST proxy with the Schema registry attached to this cluster. Requires 'integrateRestProxyWithSchemaRegistry' to be 'true'
	UseLocalSchemaRegistry bool `json:"useLocalSchemaRegistry,omitempty"`

	// URL of the Kafka schema registry to integrate with. Requires 'integrateRestProxyWithSchemaRegistry' to be 'true' and useLocalSchemaRegistry to be 'false'
	SchemaRegistryServerUrl string `json:"schemaRegistryServerUrl,omitempty"`

	// Username to use when connecting to the Kafka schema registry. Requires 'integrateRestProxyWithSchemaRegistry' to be 'true' and useLocalSchemaRegistry to be 'false'
	SchemaRegistryUsername string `json:"schemaRegistryUsername,omitempty"`

	// Password to use when connecting to the Kafka schema registry. Requires 'integrateRestProxyWithSchemaRegistry' to be 'true' and useLocalSchemaRegistry to be 'false'
	SchemaRegistryPassword string `json:"schemaRegistryPassword,omitempty"`
}

// AssertKafkaRestProxyOptionsSchemaRequired checks if the required fields are not zero-ed
func AssertKafkaRestProxyOptionsSchemaRequired(obj KafkaRestProxyOptionsSchema) error {
	return nil
}

// AssertKafkaRestProxyOptionsSchemaConstraints checks if the values respects the defined constraints
func AssertKafkaRestProxyOptionsSchemaConstraints(obj KafkaRestProxyOptionsSchema) error {
	return nil
}
