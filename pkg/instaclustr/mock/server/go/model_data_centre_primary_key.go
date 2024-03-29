/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// DataCentrePrimaryKey - DataCentre of the rack
type DataCentrePrimaryKey struct {

	// Name of data centre
	Name string `json:"name,omitempty"`

	// Provider name
	Provider string `json:"provider,omitempty"`
}

// AssertDataCentrePrimaryKeyRequired checks if the required fields are not zero-ed
func AssertDataCentrePrimaryKeyRequired(obj DataCentrePrimaryKey) error {
	return nil
}

// AssertDataCentrePrimaryKeyConstraints checks if the values respects the defined constraints
func AssertDataCentrePrimaryKeyConstraints(obj DataCentrePrimaryKey) error {
	return nil
}
