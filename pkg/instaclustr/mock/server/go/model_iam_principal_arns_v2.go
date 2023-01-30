/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// IamPrincipalArnsV2 - List of IAM Principal ARNs for a cluster data center
type IamPrincipalArnsV2 struct {

	// ID of the cluster data center
	ClusterDataCenterId string `json:"clusterDataCenterId,omitempty"`

	// IAM Principal ARNs to allow connection to the AWS Endpoint Service.
	IamPrincipalArns []IamPrincipalArnV2 `json:"iamPrincipalArns,omitempty"`
}

// AssertIamPrincipalArnsV2Required checks if the required fields are not zero-ed
func AssertIamPrincipalArnsV2Required(obj IamPrincipalArnsV2) error {
	for _, el := range obj.IamPrincipalArns {
		if err := AssertIamPrincipalArnV2Required(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertRecurseIamPrincipalArnsV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of IamPrincipalArnsV2 (e.g. [][]IamPrincipalArnsV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseIamPrincipalArnsV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aIamPrincipalArnsV2, ok := obj.(IamPrincipalArnsV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertIamPrincipalArnsV2Required(aIamPrincipalArnsV2)
	})
}
