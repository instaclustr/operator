/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type ResizeCompatibleNodeSizesV2 struct {

	// List of compatible node sizes based on the purpose of the node.
	CompatibleNodeSizes []ResizeCompatibleNodeSizeInfoV2 `json:"compatibleNodeSizes,omitempty"`
}

// AssertResizeCompatibleNodeSizesV2Required checks if the required fields are not zero-ed
func AssertResizeCompatibleNodeSizesV2Required(obj ResizeCompatibleNodeSizesV2) error {
	for _, el := range obj.CompatibleNodeSizes {
		if err := AssertResizeCompatibleNodeSizeInfoV2Required(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertResizeCompatibleNodeSizesV2Constraints checks if the values respects the defined constraints
func AssertResizeCompatibleNodeSizesV2Constraints(obj ResizeCompatibleNodeSizesV2) error {
	return nil
}
