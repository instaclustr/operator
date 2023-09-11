/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type ApiVpcPeeringConnectionRequestMixedSchema struct {

	// The account ID of the owner of the accepter VPC
	PeerAccountId string `json:"peerAccountId"`

	// ID of the VPC with which you are creating the peering connection
	PeerVpcId string `json:"peerVpcId"`

	// The subnet for the peering VPC, mutually exclusive with peerSubnets. Please note, this is a deprecated field, and the peerSubnets field should be used instead.
	PeerSubnet string `json:"peerSubnet,omitempty"`

	// The subnets for the peering VPC, mutually exclusive with peerSubnet
	PeerSubnets []string `json:"peerSubnets,omitempty"`

	// Region code for the accepter VPC, if the accepter VPC is located in a region other than the region in which you make the request.<br><br>Defaults to the region of the Cluster Data Centre.
	PeerRegion string `json:"peerRegion,omitempty"`

	// Whether to add the VPC peer network to the cluster firewall allowed addresses<br><br>Defaults to false.
	AddNetworkToFirewallRules bool `json:"addNetworkToFirewallRules,omitempty"`

	ValidationMessages map[string]string `json:"validationMessages,omitempty"`

	// Resource Group Name of the Virtual Network
	PeerResourceGroup string `json:"peerResourceGroup"`

	// Subscription Id of the Virtual Network
	PeerSubscriptionId string `json:"peerSubscriptionId"`

	//  Id of the Active Directory Object to give peering permissions to, required for cross subscription peering
	PeerAdObjectId string `json:"peerAdObjectId,omitempty"`

	// Name of the Virtual Network you wish to peer to
	PeerVNetNetworkName string `json:"peerVNetNetworkName"`

	// VPC Network Name of the VPC with which you are creating the peering connection
	PeerVpcNetworkName string `json:"peerVpcNetworkName"`

	// The project ID of the owner of the accepter VPC
	PeerProjectId string `json:"peerProjectId"`
}

// AssertApiVpcPeeringConnectionRequestMixedSchemaRequired checks if the required fields are not zero-ed
func AssertApiVpcPeeringConnectionRequestMixedSchemaRequired(obj ApiVpcPeeringConnectionRequestMixedSchema) error {
	elements := map[string]interface{}{
		"peerAccountId":       obj.PeerAccountId,
		"peerVpcId":           obj.PeerVpcId,
		"peerResourceGroup":   obj.PeerResourceGroup,
		"peerSubscriptionId":  obj.PeerSubscriptionId,
		"peerVNetNetworkName": obj.PeerVNetNetworkName,
		"peerVpcNetworkName":  obj.PeerVpcNetworkName,
		"peerProjectId":       obj.PeerProjectId,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertApiVpcPeeringConnectionRequestMixedSchemaConstraints checks if the values respects the defined constraints
func AssertApiVpcPeeringConnectionRequestMixedSchemaConstraints(obj ApiVpcPeeringConnectionRequestMixedSchema) error {
	return nil
}
