/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// AccountAzureVnetPeersV2 - A listable data source of all Azure Virtual Network Peering requests in an Instaclustr Account.
type AccountAzureVnetPeersV2 struct {

	// UUID of the Instaclustr Account.
	AccountId string `json:"accountId,omitempty"`

	PeeringRequests []AzureVnetPeerSummaryV2 `json:"peeringRequests,omitempty"`
}

// AssertAccountAzureVnetPeersV2Required checks if the required fields are not zero-ed
func AssertAccountAzureVnetPeersV2Required(obj AccountAzureVnetPeersV2) error {
	for _, el := range obj.PeeringRequests {
		if err := AssertAzureVnetPeerSummaryV2Required(el); err != nil {
			return err
		}
	}
	return nil
}

// AssertAccountAzureVnetPeersV2Constraints checks if the values respects the defined constraints
func AssertAccountAzureVnetPeersV2Constraints(obj AccountAzureVnetPeersV2) error {
	return nil
}
