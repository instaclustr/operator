/*
 * Instaclustr API Documentation
 *
 *
 *
 * API version: Current
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

// AWSVPCPeerV2APIService is a service that implements the logic for the AWSVPCPeerV2APIServicer
// This service should implement the business logic for every endpoint for the AWSVPCPeerV2API API.
// Include any external packages or services that will be required by this service.
type AWSVPCPeerV2APIService struct {
	mu    sync.RWMutex
	peers map[string]*AwsVpcPeerV2
}

// NewAWSVPCPeerV2APIService creates a default api service
func NewAWSVPCPeerV2APIService() AWSVPCPeerV2APIServicer {
	return &AWSVPCPeerV2APIService{peers: map[string]*AwsVpcPeerV2{}}
}

// ClusterManagementV2DataSourcesProvidersAwsVpcPeersV2Get - List all AWS VPC Peering requests
func (s *AWSVPCPeerV2APIService) ClusterManagementV2DataSourcesProvidersAwsVpcPeersV2Get(ctx context.Context) (ImplResponse, error) {
	// TODO - update ClusterManagementV2DataSourcesProvidersAwsVpcPeersV2Get with the required logic for this service method.
	// Add api_awsvpc_peer_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, []AccountAwsVpcPeersV2{}) or use other options such as http.Ok ...
	// return Response(200, []AccountAwsVpcPeersV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2DataSourcesProvidersAwsVpcPeersV2Get method not implemented")
}

// ClusterManagementV2ResourcesProvidersAwsVpcPeersV2Post - Create AWS VPC Peering Request
func (s *AWSVPCPeerV2APIService) ClusterManagementV2ResourcesProvidersAwsVpcPeersV2Post(ctx context.Context, awsVpcPeerV2 AwsVpcPeerV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesProvidersAwsVpcPeersV2Post with the required logic for this service method.
	// Add api_awsvpc_peer_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(202, AwsVpcPeerV2{}) or use other options such as http.Ok ...
	// return Response(202, AwsVpcPeerV2{}), nil

	s.mu.Lock()
	defer s.mu.Unlock()

	id, err := uuid.NewUUID()
	if err != nil {
		return Response(http.StatusInternalServerError, nil), err
	}

	awsVpcPeerV2.Id = id.String()

	s.peers[awsVpcPeerV2.Id] = &awsVpcPeerV2

	return Response(http.StatusAccepted, &awsVpcPeerV2), nil
}

// ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdDelete - Delete AWS VPC Peering Connection
func (s *AWSVPCPeerV2APIService) ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdDelete(ctx context.Context, vpcPeerId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdDelete with the required logic for this service method.
	// Add api_awsvpc_peer_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(204, {}) or use other options such as http.Ok ...
	// return Response(204, nil),nil

	// TODO: Uncomment the next line to return response Response(404, ErrorListResponseV2{}) or use other options such as http.Ok ...
	// return Response(404, ErrorListResponseV2{}), nil

	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.peers[vpcPeerId]
	if !exists {
		return Response(http.StatusNotFound, nil), nil
	}

	delete(s.peers, vpcPeerId)

	return Response(http.StatusNoContent, nil), nil
}

// ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdGet - Get AWS VPC Peering Connection info
func (s *AWSVPCPeerV2APIService) ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdGet(ctx context.Context, vpcPeerId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdGet with the required logic for this service method.
	// Add api_awsvpc_peer_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, AwsVpcPeerV2{}) or use other options such as http.Ok ...
	// return Response(200, AwsVpcPeerV2{}), nil

	// TODO: Uncomment the next line to return response Response(404, ErrorListResponseV2{}) or use other options such as http.Ok ...
	// return Response(404, ErrorListResponseV2{}), nil

	s.mu.Lock()
	defer s.mu.Unlock()

	peer, exists := s.peers[vpcPeerId]
	if !exists {
		return Response(http.StatusNotFound, nil), nil
	}

	return Response(http.StatusOK, peer), nil
}

// ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdPut - Update AWS VPC Peering Connection info
func (s *AWSVPCPeerV2APIService) ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdPut(ctx context.Context, vpcPeerId string, awsVpcPeerUpdateV2 AwsVpcPeerUpdateV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdPut with the required logic for this service method.
	// Add api_awsvpc_peer_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(202, AwsVpcPeerV2{}) or use other options such as http.Ok ...
	// return Response(202, AwsVpcPeerV2{}), nil

	// TODO: Uncomment the next line to return response Response(404, ErrorListResponseV2{}) or use other options such as http.Ok ...
	// return Response(404, ErrorListResponseV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesProvidersAwsVpcPeersV2VpcPeerIdPut method not implemented")
}
