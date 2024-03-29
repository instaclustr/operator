/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type ApiVpcPeeringConnectionMixedSchema struct {

	// ID of the VPC peering connection
	Id string `json:"id,omitempty"`

	// ID of VPC peering connection on AWS
	AwsVpcConnectionId string `json:"aws_vpc_connection_id,omitempty"`

	// ID of the Cluster Data Centre
	ClusterDataCentre string `json:"clusterDataCentre,omitempty"`

	// AWS VPC ID for the Cluster Data Centre
	VpcId string `json:"vpcId,omitempty"`

	// ID of the VPC with which you are creating the peering connection
	PeerVpcId string `json:"peerVpcId,omitempty"`

	// The account ID of the owner of the accepter VPC
	PeerAccountId string `json:"peerAccountId,omitempty"`

	PeerSubnet CidrSchema `json:"peerSubnet,omitempty"`

	// Region code for the accepter VPC, if the accepter VPC is located in a region other than the region in which you make the request.
	PeerRegion string `json:"peerRegion,omitempty"`

	PeerSubnets AzureApiPeeringConnectionPeerSubnets `json:"peerSubnets,omitempty"`

	// Name of the Vpc Peering Connection
	Name string `json:"name,omitempty"`

	// Subscription Id of the peered Virtual Network
	PeerSubscriptionId string `json:"peerSubscriptionId,omitempty"`

	// ID of the Active Directory Object to give peering permissions to, required for cross subscription peering
	PeerAdObjectId string `json:"peerAdObjectId,omitempty"`

	// Resource Group name of the peered Virtual Network
	PeerResourceGroup string `json:"peerResourceGroup,omitempty"`

	// Name of the peered Virtual Network
	PeerVNet string `json:"peerVNet,omitempty"`

	// Subscription Id of the of the Datacenter Virtual Network
	SubscriptionId string `json:"subscriptionId,omitempty"`

	// Resource Group name of the Datacenter Virtual Network
	ResourceGroup string `json:"resourceGroup,omitempty"`

	// Name of the Datacenter Virtual Network
	VNet string `json:"vNet,omitempty"`

	// Reason for Peering Connection Failure
	FailureReason string `json:"failureReason,omitempty"`

	// Vpc Network Name of the Datacenter VPC
	VpcNetworkName string `json:"vpcNetworkName,omitempty"`

	// Vpc Network Name of the Peered VPC
	PeerVpcNetworkName string `json:"peerVpcNetworkName,omitempty"`

	// GCP Project ID of the Datacenter
	ProjectId string `json:"projectId,omitempty"`

	// GCP Project ID of the Peered VPC
	PeerProjectId string `json:"peerProjectId,omitempty"`

	// Status of the VPC peering connection
	StatusCode string `json:"statusCode,omitempty"`
}

// AssertApiVpcPeeringConnectionMixedSchemaRequired checks if the required fields are not zero-ed
func AssertApiVpcPeeringConnectionMixedSchemaRequired(obj ApiVpcPeeringConnectionMixedSchema) error {
	if err := AssertCidrSchemaRequired(obj.PeerSubnet); err != nil {
		return err
	}
	if err := AssertAzureApiPeeringConnectionPeerSubnetsRequired(obj.PeerSubnets); err != nil {
		return err
	}
	return nil
}

// AssertApiVpcPeeringConnectionMixedSchemaConstraints checks if the values respects the defined constraints
func AssertApiVpcPeeringConnectionMixedSchemaConstraints(obj ApiVpcPeeringConnectionMixedSchema) error {
	return nil
}
