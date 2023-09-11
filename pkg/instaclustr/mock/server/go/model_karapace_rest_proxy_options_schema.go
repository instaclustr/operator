/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type KarapaceRestProxyOptionsSchema struct {

	// Accepts true/false. Enables Integration of the Karapace REST Proxy with the local Karapace Schema Registry
	IntegrateRestProxyWithSchemaRegistry bool `json:"integrateRestProxyWithSchemaRegistry,omitempty"`
}

// AssertKarapaceRestProxyOptionsSchemaRequired checks if the required fields are not zero-ed
func AssertKarapaceRestProxyOptionsSchemaRequired(obj KarapaceRestProxyOptionsSchema) error {
	return nil
}

// AssertKarapaceRestProxyOptionsSchemaConstraints checks if the values respects the defined constraints
func AssertKarapaceRestProxyOptionsSchemaConstraints(obj KarapaceRestProxyOptionsSchema) error {
	return nil
}
