/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type ApiEncryptionKeyRequest struct {

	// Encryption key alias
	Alias string `json:"alias"`

	// AWS ARN for the encryption key
	Arn string `json:"arn"`

	Provider string `json:"provider,omitempty"`

	// Name of the provider account. Defaults to <code>INSTACLUSTR</code>. Must be explicitly provided for Run-In-Your-Own-Account customers.
	ProviderAccount string `json:"providerAccount,omitempty"`
}

// AssertApiEncryptionKeyRequestRequired checks if the required fields are not zero-ed
func AssertApiEncryptionKeyRequestRequired(obj ApiEncryptionKeyRequest) error {
	elements := map[string]interface{}{
		"alias": obj.Alias,
		"arn":   obj.Arn,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertApiEncryptionKeyRequestConstraints checks if the values respects the defined constraints
func AssertApiEncryptionKeyRequestConstraints(obj ApiEncryptionKeyRequest) error {
	return nil
}
