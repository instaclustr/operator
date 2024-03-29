/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type NodeDetailsV2 struct {

	// The roles or purposes of the node. Useful for filtering for nodes that have a specific role.
	NodeRoles []NodeRolesV2 `json:"nodeRoles,omitempty"`

	// Rack name in which the node is located.
	Rack string `json:"rack,omitempty"`

	// Size of the node.
	NodeSize string `json:"nodeSize,omitempty"`

	// Private IP address of the node.
	PrivateAddress string `json:"privateAddress,omitempty"`

	// Deletion time of the node as a UTC timestamp
	DeletionTime string `json:"deletionTime,omitempty"`

	// Public IP address of the node.
	PublicAddress string `json:"publicAddress,omitempty"`

	// Start time of the node as a UTC timestamp
	StartTime string `json:"startTime,omitempty"`

	// ID of the node.
	Id string `json:"id,omitempty"`

	// Provisioning status of the node.
	Status string `json:"status,omitempty"`
}

// AssertNodeDetailsV2Required checks if the required fields are not zero-ed
func AssertNodeDetailsV2Required(obj NodeDetailsV2) error {
	return nil
}

// AssertNodeDetailsV2Constraints checks if the values respects the defined constraints
func AssertNodeDetailsV2Constraints(obj NodeDetailsV2) error {
	return nil
}
