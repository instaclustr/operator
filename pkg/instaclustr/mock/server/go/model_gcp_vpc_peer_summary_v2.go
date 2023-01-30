/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// GcpVpcPeerSummaryV2 -
type GcpVpcPeerSummaryV2 struct {

	// The project ID of the owner of the accepter VPC.
	PeerProjectId string `json:"peerProjectId"`

	// The subnets for the peering VPC.
	PeerSubnets []string `json:"peerSubnets"`

	// Name of the Peering Connection.
	Name string `json:"name,omitempty"`

	// ID of the VPC peering connection.
	Id string `json:"id,omitempty"`

	// The name of the VPC Network you wish to peer to.
	PeerVpcNetworkName string `json:"peerVpcNetworkName"`

	// ID of the Cluster Data Centre.
	CdcId string `json:"cdcId"`
}

// AssertGcpVpcPeerSummaryV2Required checks if the required fields are not zero-ed
func AssertGcpVpcPeerSummaryV2Required(obj GcpVpcPeerSummaryV2) error {
	elements := map[string]interface{}{
		"peerProjectId":      obj.PeerProjectId,
		"peerSubnets":        obj.PeerSubnets,
		"peerVpcNetworkName": obj.PeerVpcNetworkName,
		"cdcId":              obj.CdcId,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertRecurseGcpVpcPeerSummaryV2Required recursively checks if required fields are not zero-ed in a nested slice.
// Accepts only nested slice of GcpVpcPeerSummaryV2 (e.g. [][]GcpVpcPeerSummaryV2), otherwise ErrTypeAssertionError is thrown.
func AssertRecurseGcpVpcPeerSummaryV2Required(objSlice interface{}) error {
	return AssertRecurseInterfaceRequired(objSlice, func(obj interface{}) error {
		aGcpVpcPeerSummaryV2, ok := obj.(GcpVpcPeerSummaryV2)
		if !ok {
			return ErrTypeAssertionError
		}
		return AssertGcpVpcPeerSummaryV2Required(aGcpVpcPeerSummaryV2)
	})
}
