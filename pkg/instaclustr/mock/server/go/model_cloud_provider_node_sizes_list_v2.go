/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// CloudProviderNodeSizesListV2 - Object containing list of node sizes grouped by cloud provider.
type CloudProviderNodeSizesListV2 struct {

	// List of all compatible cloud provider node sizes.
	CloudProviderNodeSizes []CloudProviderNodeSizesV2 `json:"cloudProviderNodeSizes,omitempty"`
}

// AssertCloudProviderNodeSizesListV2Required checks if the required fields are not zero-ed
func AssertCloudProviderNodeSizesListV2Required(obj CloudProviderNodeSizesListV2) error {
	for _, el := range obj.CloudProviderNodeSizes {
		if err := AssertCloudProviderNodeSizesV2Required(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertCloudProviderNodeSizesListV2Constraints checks if the values respects the defined constraints
func AssertCloudProviderNodeSizesListV2Constraints(obj CloudProviderNodeSizesListV2) error {
	return nil
}