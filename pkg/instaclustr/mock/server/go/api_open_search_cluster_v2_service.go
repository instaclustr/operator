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

// OpenSearchClusterV2ApiService is a service that implements the logic for the OpenSearchClusterV2ApiServicer
// This service should implement the business logic for every endpoint for the OpenSearchClusterV2Api API.
// Include any external packages or services that will be required by this service.
type OpenSearchClusterV2ApiService struct {
}

// NewOpenSearchClusterV2ApiService creates a default api service
func NewOpenSearchClusterV2ApiService() OpenSearchClusterV2ApiServicer {
	return &OpenSearchClusterV2ApiService{}
}

// ClusterManagementV2ResourcesApplicationsOpensearchClustersV2ClusterIdDelete - Delete cluster
func (s *OpenSearchClusterV2ApiService) ClusterManagementV2ResourcesApplicationsOpensearchClustersV2ClusterIdDelete(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsOpensearchClustersV2ClusterIdDelete with the required logic for this service method.
	// Add api_open_search_cluster_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(204, {}) or use other options such as http.Ok ...
	//return Response(204, nil),nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesApplicationsOpensearchClustersV2ClusterIdDelete method not implemented")
}

// ClusterManagementV2ResourcesApplicationsOpensearchClustersV2ClusterIdGet - Get OpenSearch cluster details
func (s *OpenSearchClusterV2ApiService) ClusterManagementV2ResourcesApplicationsOpensearchClustersV2ClusterIdGet(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsOpensearchClustersV2ClusterIdGet with the required logic for this service method.
	// Add api_open_search_cluster_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, OpenSearchClusterV2{}) or use other options such as http.Ok ...
	//return Response(200, OpenSearchClusterV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesApplicationsOpensearchClustersV2ClusterIdGet method not implemented")
}

// ClusterManagementV2ResourcesApplicationsOpensearchClustersV2Post - Create an OpenSearch cluster
func (s *OpenSearchClusterV2ApiService) ClusterManagementV2ResourcesApplicationsOpensearchClustersV2Post(ctx context.Context, body OpenSearchClusterV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsOpensearchClustersV2Post with the required logic for this service method.
	// Add api_open_search_cluster_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(202, OpenSearchClusterV2{}) or use other options such as http.Ok ...
	//return Response(202, OpenSearchClusterV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesApplicationsOpensearchClustersV2Post method not implemented")
}
