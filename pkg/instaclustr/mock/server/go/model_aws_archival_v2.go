/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// AwsArchivalV2 -
type AwsArchivalV2 struct {

	// S3 resource region
	ArchivalS3Region string `json:"archivalS3Region"`

	// AWS access Key ID
	AwsAccessKeyId string `json:"awsAccessKeyId,omitempty"`

	// S3 resource URI
	ArchivalS3Uri string `json:"archivalS3Uri"`

	// AWS secret access key
	AwsSecretAccessKey string `json:"awsSecretAccessKey,omitempty"`
}

// AssertAwsArchivalV2Required checks if the required fields are not zero-ed
func AssertAwsArchivalV2Required(obj AwsArchivalV2) error {
	elements := map[string]interface{}{
		"archivalS3Region": obj.ArchivalS3Region,
		"archivalS3Uri":    obj.ArchivalS3Uri,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertRecurseAwsArchivalV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of AwsArchivalV2 (e.g. [][]AwsArchivalV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseAwsArchivalV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aAwsArchivalV2, ok := obj.(AwsArchivalV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertAwsArchivalV2Required(aAwsArchivalV2)
	})
}