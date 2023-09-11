/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type AwsEndpointServicePrincipalsRequest struct {

	// List of IAM Principal ARNs
	IamPrincipalARNs []string `json:"iamPrincipalARNs,omitempty"`

	ValidationMessages map[string]string `json:"validationMessages,omitempty"`
}

// AssertAwsEndpointServicePrincipalsRequestRequired checks if the required fields are not zero-ed
func AssertAwsEndpointServicePrincipalsRequestRequired(obj AwsEndpointServicePrincipalsRequest) error {
	return nil
}

// AssertAwsEndpointServicePrincipalsRequestConstraints checks if the values respects the defined constraints
func AssertAwsEndpointServicePrincipalsRequestConstraints(obj AwsEndpointServicePrincipalsRequest) error {
	return nil
}
