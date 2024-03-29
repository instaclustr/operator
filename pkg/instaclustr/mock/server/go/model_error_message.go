/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type ErrorMessage struct {
	Status int32 `json:"status,omitempty"`

	Message string `json:"message,omitempty"`

	ErrorDetails map[string]interface{} `json:"errorDetails,omitempty"`

	Code int32 `json:"code,omitempty"`

	Link string `json:"link,omitempty"`
}

// AssertErrorMessageRequired checks if the required fields are not zero-ed
func AssertErrorMessageRequired(obj ErrorMessage) error {
	return nil
}

// AssertErrorMessageConstraints checks if the values respects the defined constraints
func AssertErrorMessageConstraints(obj ErrorMessage) error {
	return nil
}
