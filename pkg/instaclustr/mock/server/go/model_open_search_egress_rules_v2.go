/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// OpenSearchEgressRulesV2 - List of OpenSearch egress rules
type OpenSearchEgressRulesV2 struct {

	// List of OpenSearch egress rules
	Rules []OpenSearchEgressRuleV2 `json:"rules"`

	// OpenSearch cluster id
	ClusterId string `json:"clusterId,omitempty"`
}

// AssertOpenSearchEgressRulesV2Required checks if the required fields are not zero-ed
func AssertOpenSearchEgressRulesV2Required(obj OpenSearchEgressRulesV2) error {
	elements := map[string]interface{}{
		"rules": obj.Rules,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	for _, el := range obj.Rules {
		if err := AssertOpenSearchEgressRuleV2Required(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertOpenSearchEgressRulesV2Constraints checks if the values respects the defined constraints
func AssertOpenSearchEgressRulesV2Constraints(obj OpenSearchEgressRulesV2) error {
	return nil
}
