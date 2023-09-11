/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type AwsProviderEncryptionKey struct {
	Id string `json:"id,omitempty"`

	Alias string `json:"alias,omitempty"`

	Arn AwsKmsArn `json:"arn,omitempty"`

	AccountPrimaryKey string `json:"accountPrimaryKey,omitempty"`

	InUse bool `json:"inUse,omitempty"`

	ProviderAccount string `json:"providerAccount,omitempty"`
}

// AssertAwsProviderEncryptionKeyRequired checks if the required fields are not zero-ed
func AssertAwsProviderEncryptionKeyRequired(obj AwsProviderEncryptionKey) error {
	if err := AssertAwsKmsArnRequired(obj.Arn); err != nil {
		return err
	}
	return nil
}

// AssertAwsProviderEncryptionKeyConstraints checks if the values respects the defined constraints
func AssertAwsProviderEncryptionKeyConstraints(obj AwsProviderEncryptionKey) error {
	return nil
}
