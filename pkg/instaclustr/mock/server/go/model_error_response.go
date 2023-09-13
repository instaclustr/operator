/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type ErrorResponse struct {
	Error string `json:"error,omitempty"`
}

// AssertErrorResponseRequired checks if the required fields are not zero-ed
func AssertErrorResponseRequired(obj ErrorResponse) error {
	return nil
}

// AssertErrorResponseConstraints checks if the values respects the defined constraints
func AssertErrorResponseConstraints(obj ErrorResponse) error {
	return nil
}