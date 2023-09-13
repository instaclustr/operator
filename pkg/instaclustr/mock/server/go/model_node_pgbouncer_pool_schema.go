/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type NodePgbouncerPoolSchema struct {
	NodeId string `json:"nodeId,omitempty"`

	Pools []PgbouncerPoolSchema `json:"pools,omitempty"`
}

// AssertNodePgbouncerPoolSchemaRequired checks if the required fields are not zero-ed
func AssertNodePgbouncerPoolSchemaRequired(obj NodePgbouncerPoolSchema) error {
	for _, el := range obj.Pools {
		if err := AssertPgbouncerPoolSchemaRequired(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertNodePgbouncerPoolSchemaConstraints checks if the values respects the defined constraints
func AssertNodePgbouncerPoolSchemaConstraints(obj NodePgbouncerPoolSchema) error {
	return nil
}