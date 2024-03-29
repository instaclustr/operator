/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// ProviderAccountPrimaryKey - Provider information including name and account
type ProviderAccountPrimaryKey struct {

	// Provider account
	Name string `json:"name,omitempty"`

	// Provider name
	Provider string `json:"provider,omitempty"`
}

// AssertProviderAccountPrimaryKeyRequired checks if the required fields are not zero-ed
func AssertProviderAccountPrimaryKeyRequired(obj ProviderAccountPrimaryKey) error {
	return nil
}

// AssertProviderAccountPrimaryKeyConstraints checks if the values respects the defined constraints
func AssertProviderAccountPrimaryKeyConstraints(obj ProviderAccountPrimaryKey) error {
	return nil
}
