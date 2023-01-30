/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// ProviderAzureSettingsV2 -
type ProviderAzureSettingsV2 struct {

	// The name of the Azure Resource Group into which the Data Centre will be provisioned.
	ResourceGroup string `json:"resourceGroup,omitempty"`
}

// AssertProviderAzureSettingsV2Required checks if the required fields are not zero-ed
func AssertProviderAzureSettingsV2Required(obj ProviderAzureSettingsV2) error {
	return nil
}

// AssertRecurseProviderAzureSettingsV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of ProviderAzureSettingsV2 (e.g. [][]ProviderAzureSettingsV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseProviderAzureSettingsV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aProviderAzureSettingsV2, ok := obj.(ProviderAzureSettingsV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertProviderAzureSettingsV2Required(aProviderAzureSettingsV2)
	})
}
