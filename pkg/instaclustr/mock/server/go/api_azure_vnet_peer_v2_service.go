/*
 * Instaclustr Cluster Management API
 *
 * Instaclustr Cluster Management API
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

import (
	"context"
	"errors"
	"net/http"
)

// AzureVnetPeerV2ApiService is a service that implements the logic for the AzureVnetPeerV2ApiServicer
// This service should implement the business logic for every endpoint for the AzureVnetPeerV2Api API.
// Include any external packages or services that will be required by this service.
type AzureVnetPeerV2ApiService struct {
}

// NewAzureVnetPeerV2ApiService creates a default api service
func NewAzureVnetPeerV2ApiService() AzureVnetPeerV2ApiServicer {
	return &AzureVnetPeerV2ApiService{}
}

// ClusterManagementV2DataSourcesProvidersAzureVnetPeersV2Get - List all Azure VNet Peering requests
func (s *AzureVnetPeerV2ApiService) ClusterManagementV2DataSourcesProvidersAzureVnetPeersV2Get(ctx context.Context) (ImplResponse, error) {
	// TODO - update ClusterManagementV2DataSourcesProvidersAzureVnetPeersV2Get with the required logic for this service method.
	// Add api_azure_vnet_peer_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, []AccountAzureVnetPeersV2{}) or use other options such as http.Ok ...
	//return Response(200, []AccountAzureVnetPeersV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2DataSourcesProvidersAzureVnetPeersV2Get method not implemented")
}

// ClusterManagementV2ResourcesProvidersAzureVnetPeersV2Post - Create Azure Vnet Peering Request
func (s *AzureVnetPeerV2ApiService) ClusterManagementV2ResourcesProvidersAzureVnetPeersV2Post(ctx context.Context, body AzureVnetPeerV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesProvidersAzureVnetPeersV2Post with the required logic for this service method.
	// Add api_azure_vnet_peer_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(202, AzureVnetPeerV2{}) or use other options such as http.Ok ...
	//return Response(202, AzureVnetPeerV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesProvidersAzureVnetPeersV2Post method not implemented")
}

// ClusterManagementV2ResourcesProvidersAzureVnetPeersV2VpcPeerIdDelete - Delete Azure Vnet Peering Connection
func (s *AzureVnetPeerV2ApiService) ClusterManagementV2ResourcesProvidersAzureVnetPeersV2VpcPeerIdDelete(ctx context.Context, vpcPeerId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesProvidersAzureVnetPeersV2VpcPeerIdDelete with the required logic for this service method.
	// Add api_azure_vnet_peer_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(204, {}) or use other options such as http.Ok ...
	//return Response(204, nil),nil

	//TODO: Uncomment the next line to return response Response(404, ErrorListResponseV2{}) or use other options such as http.Ok ...
	//return Response(404, ErrorListResponseV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesProvidersAzureVnetPeersV2VpcPeerIdDelete method not implemented")
}

// ClusterManagementV2ResourcesProvidersAzureVnetPeersV2VpcPeerIdGet - Get Azure Vnet Peering Connection info
func (s *AzureVnetPeerV2ApiService) ClusterManagementV2ResourcesProvidersAzureVnetPeersV2VpcPeerIdGet(ctx context.Context, vpcPeerId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesProvidersAzureVnetPeersV2VpcPeerIdGet with the required logic for this service method.
	// Add api_azure_vnet_peer_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, AzureVnetPeerV2{}) or use other options such as http.Ok ...
	//return Response(200, AzureVnetPeerV2{}), nil

	//TODO: Uncomment the next line to return response Response(404, ErrorListResponseV2{}) or use other options such as http.Ok ...
	//return Response(404, ErrorListResponseV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesProvidersAzureVnetPeersV2VpcPeerIdGet method not implemented")
}