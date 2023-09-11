/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type GenericJsonErrorListResponse struct {
	Errors []GenericJsonErrorResponse `json:"errors,omitempty"`
}

// AssertGenericJsonErrorListResponseRequired checks if the required fields are not zero-ed
func AssertGenericJsonErrorListResponseRequired(obj GenericJsonErrorListResponse) error {
	for _, el := range obj.Errors {
		if err := AssertGenericJsonErrorResponseRequired(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertGenericJsonErrorListResponseConstraints checks if the values respects the defined constraints
func AssertGenericJsonErrorListResponseConstraints(obj GenericJsonErrorListResponse) error {
	return nil
}
