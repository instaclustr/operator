/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type DeleteNodesOperationInfoV2 struct {

	// Number of nodes set to delete in the operation.
	NumberOfNodesToDelete int32 `json:"numberOfNodesToDelete,omitempty"`

	DeleteNodeOperations []DeleteNodeOperationInfoV2 `json:"deleteNodeOperations,omitempty"`

	// Timestamp of the creation of the operation
	Created string `json:"created,omitempty"`

	// Timestamp of the last modification of the operation
	Modified string `json:"modified,omitempty"`

	// Operation id
	Id string `json:"id,omitempty"`

	// ID of the Cluster Data Centre.
	CdcId string `json:"cdcId,omitempty"`

	Status OperationStatusV2 `json:"status,omitempty"`
}

// AssertDeleteNodesOperationInfoV2Required checks if the required fields are not zero-ed
func AssertDeleteNodesOperationInfoV2Required(obj DeleteNodesOperationInfoV2) error {
	for _, el := range obj.DeleteNodeOperations {
		if err := AssertDeleteNodeOperationInfoV2Required(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertDeleteNodesOperationInfoV2Constraints checks if the values respects the defined constraints
func AssertDeleteNodesOperationInfoV2Constraints(obj DeleteNodesOperationInfoV2) error {
	return nil
}
