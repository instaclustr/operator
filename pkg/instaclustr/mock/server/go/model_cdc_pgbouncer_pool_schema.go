/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type CdcPgbouncerPoolSchema struct {
	CdcId string `json:"cdcId,omitempty"`

	Nodes []NodePgbouncerPoolSchema `json:"nodes,omitempty"`
}

// AssertCdcPgbouncerPoolSchemaRequired checks if the required fields are not zero-ed
func AssertCdcPgbouncerPoolSchemaRequired(obj CdcPgbouncerPoolSchema) error {
	for _, el := range obj.Nodes {
		if err := AssertNodePgbouncerPoolSchemaRequired(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertCdcPgbouncerPoolSchemaConstraints checks if the values respects the defined constraints
func AssertCdcPgbouncerPoolSchemaConstraints(obj CdcPgbouncerPoolSchema) error {
	return nil
}