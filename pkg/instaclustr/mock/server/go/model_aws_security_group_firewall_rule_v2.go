/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// AwsSecurityGroupFirewallRuleV2 - Definition of an AWS Security Group based firewall rule to be applied to a cluster.
type AwsSecurityGroupFirewallRuleV2 struct {

	// The security group ID of the AWS security group firewall rule.
	SecurityGroupId string `json:"securityGroupId"`

	// The reason (if needed) for the deferred status of the AWS security group firewall rule.
	DeferredReason string `json:"deferredReason,omitempty"`

	// ID of the cluster for the AWS security group firewall rule.
	ClusterId string `json:"clusterId"`

	// ID of the AWS security group firewall rule.
	Id string `json:"id,omitempty"`

	Type FirewallRuleTypesV2 `json:"type"`

	// The status of the AWS security group firewall rule.
	Status string `json:"status,omitempty"`
}

// AssertAwsSecurityGroupFirewallRuleV2Required checks if the required fields are not zero-ed
func AssertAwsSecurityGroupFirewallRuleV2Required(obj AwsSecurityGroupFirewallRuleV2) error {
	elements := map[string]interface{}{
		"securityGroupId": obj.SecurityGroupId,
		"clusterId":       obj.ClusterId,
		"type":            obj.Type,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertAwsSecurityGroupFirewallRuleV2Constraints checks if the values respects the defined constraints
func AssertAwsSecurityGroupFirewallRuleV2Constraints(obj AwsSecurityGroupFirewallRuleV2) error {
	return nil
}
