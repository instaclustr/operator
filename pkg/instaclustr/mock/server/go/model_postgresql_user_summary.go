/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// PostgresqlUserSummary - Summary of the default PostgreSQL user to be applied to a PostgreSQL cluster.
type PostgresqlUserSummary struct {

	// Password of the default PostgreSQL user.
	Password string `json:"password"`
}

// AssertPostgresqlUserSummaryRequired checks if the required fields are not zero-ed
func AssertPostgresqlUserSummaryRequired(obj PostgresqlUserSummary) error {
	elements := map[string]interface{}{
		"password": obj.Password,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertRecursePostgresqlUserSummaryRequired recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of PostgresqlUserSummary (e.g. [][]PostgresqlUserSummary), otherwise ErrTypeAssertionError is thrown.
func AssertRecursePostgresqlUserSummaryRequired(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aPostgresqlUserSummary, ok := obj.(PostgresqlUserSummary)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertPostgresqlUserSummaryRequired(aPostgresqlUserSummary)
	})
}