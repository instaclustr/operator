/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// ProviderType - The type of Application. Example: APACHE_CASSANDRA.
type ProviderType struct {
	PROVIDER DataCentreRegion `json:"PROVIDER,omitempty"`
}

// AssertProviderTypeRequired checks if the required fields are not zero-ed
func AssertProviderTypeRequired(obj ProviderType) error {
	if err := AssertDataCentreRegionRequired(obj.PROVIDER); err != nil {
		return err
	}
	return nil
}

// AssertProviderTypeConstraints checks if the values respects the defined constraints
func AssertProviderTypeConstraints(obj ProviderType) error {
	return nil
}