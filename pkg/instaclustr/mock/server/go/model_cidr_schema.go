/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type CidrSchema struct {
	Network string `json:"network,omitempty"`

	PrefixLength int32 `json:"prefixLength,omitempty"`
}

// AssertCidrSchemaRequired checks if the required fields are not zero-ed
func AssertCidrSchemaRequired(obj CidrSchema) error {
	return nil
}

// AssertCidrSchemaConstraints checks if the values respects the defined constraints
func AssertCidrSchemaConstraints(obj CidrSchema) error {
	return nil
}
