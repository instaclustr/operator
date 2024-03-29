/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// CadenceStandardProvisioningV2 - Settings for STARDARD provisioning
type CadenceStandardProvisioningV2 struct {

	// Cadence advanced visibility settings
	AdvancedVisibility []CadenceAdvancedVisibilityV2 `json:"advancedVisibility,omitempty"`

	TargetCassandra CadenceDependencyV2 `json:"targetCassandra"`
}

// AssertCadenceStandardProvisioningV2Required checks if the required fields are not zero-ed
func AssertCadenceStandardProvisioningV2Required(obj CadenceStandardProvisioningV2) error {
	elements := map[string]interface{}{
		"targetCassandra": obj.TargetCassandra,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	for _, el := range obj.AdvancedVisibility {
		if err := AssertCadenceAdvancedVisibilityV2Required(el); err != nil {
			return err
		}
	}
	if err := AssertCadenceDependencyV2Required(obj.TargetCassandra); err != nil {
		return err
	}
	return nil
}

// AssertCadenceStandardProvisioningV2Constraints checks if the values respects the defined constraints
func AssertCadenceStandardProvisioningV2Constraints(obj CadenceStandardProvisioningV2) error {
	return nil
}
