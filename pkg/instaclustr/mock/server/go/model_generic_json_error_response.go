/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type GenericJsonErrorResponse struct {
	Name string `json:"name,omitempty"`

	Message string `json:"message,omitempty"`

	ErrorClass string `json:"errorClass,omitempty"`

	Help string `json:"help,omitempty"`
}

// AssertGenericJsonErrorResponseRequired checks if the required fields are not zero-ed
func AssertGenericJsonErrorResponseRequired(obj GenericJsonErrorResponse) error {
	return nil
}

// AssertGenericJsonErrorResponseConstraints checks if the values respects the defined constraints
func AssertGenericJsonErrorResponseConstraints(obj GenericJsonErrorResponse) error {
	return nil
}
