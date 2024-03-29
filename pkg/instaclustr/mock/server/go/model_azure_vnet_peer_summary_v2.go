/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type AzureVnetPeerSummaryV2 struct {

	// Resource Group Name of the Data Centre Virtual Network.
	DataCentreResourceGroup string `json:"dataCentreResourceGroup,omitempty"`

	// The name of the VPC Network you wish to peer to.
	PeerVirtualNetworkName string `json:"peerVirtualNetworkName"`

	// The subnets for the peering VPC.
	PeerSubnets []string `json:"peerSubnets"`

	// Name of the Vpc Peering Connection.
	Name string `json:"name,omitempty"`

	// ID of the Active Directory Object to give peering permissions to, required for cross subscription peering.
	PeerAdObjectId string `json:"peerAdObjectId,omitempty"`

	// ID of the VPC peering connection.
	Id string `json:"id,omitempty"`

	// Subscription ID of the Data Centre Virtual Network.
	DataCentreSubscriptionId string `json:"dataCentreSubscriptionId,omitempty"`

	// Resource Group Name of the Virtual Network.
	PeerResourceGroup string `json:"peerResourceGroup"`

	// Subscription ID of the Virtual Network.
	PeerSubscriptionId string `json:"peerSubscriptionId"`

	// ID of the Cluster Data Centre.
	CdcId string `json:"cdcId"`

	// The name of the Data Centre Virtual Network.
	DataCentreVirtualNetworkName string `json:"dataCentreVirtualNetworkName,omitempty"`
}

// AssertAzureVnetPeerSummaryV2Required checks if the required fields are not zero-ed
func AssertAzureVnetPeerSummaryV2Required(obj AzureVnetPeerSummaryV2) error {
	elements := map[string]interface{}{
		"peerVirtualNetworkName": obj.PeerVirtualNetworkName,
		"peerSubnets":            obj.PeerSubnets,
		"peerResourceGroup":      obj.PeerResourceGroup,
		"peerSubscriptionId":     obj.PeerSubscriptionId,
		"cdcId":                  obj.CdcId,
	}
	for name, el := range elements {
		if isZero := IsZeroValue(el); isZero {
			return &RequiredError{Field: name}
		}
	}

	return nil
}

// AssertAzureVnetPeerSummaryV2Constraints checks if the values respects the defined constraints
func AssertAzureVnetPeerSummaryV2Constraints(obj AzureVnetPeerSummaryV2) error {
	return nil
}
