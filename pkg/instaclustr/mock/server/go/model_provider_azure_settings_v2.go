/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type ProviderAzureSettingsV2 struct {

	// The name of the Azure Resource Group into which the Data Centre will be provisioned.
	ResourceGroup string `json:"resourceGroup,omitempty"`
}

// AssertProviderAzureSettingsV2Required checks if the required fields are not zero-ed
func AssertProviderAzureSettingsV2Required(obj ProviderAzureSettingsV2) error {
	return nil
}

// AssertProviderAzureSettingsV2Constraints checks if the values respects the defined constraints
func AssertProviderAzureSettingsV2Constraints(obj ProviderAzureSettingsV2) error {
	return nil
}
