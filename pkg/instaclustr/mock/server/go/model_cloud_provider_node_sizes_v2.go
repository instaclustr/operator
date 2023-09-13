/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// CloudProviderNodeSizesV2 - Object containing list of compatible node sizes in a cloud provider.
type CloudProviderNodeSizesV2 struct {

	// List of node size names.
	NodeSizes []string `json:"nodeSizes,omitempty"`

	CloudProvider CloudProviderEnumV2 `json:"cloudProvider,omitempty"`

	// Cloud provider documentation of the virtual machine instances.
	Document string `json:"document,omitempty"`
}

// AssertCloudProviderNodeSizesV2Required checks if the required fields are not zero-ed
func AssertCloudProviderNodeSizesV2Required(obj CloudProviderNodeSizesV2) error {
	return nil
}

// AssertCloudProviderNodeSizesV2Constraints checks if the values respects the defined constraints
func AssertCloudProviderNodeSizesV2Constraints(obj CloudProviderNodeSizesV2) error {
	return nil
}