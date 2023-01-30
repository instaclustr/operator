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

// ApacheZookeeperClusterV2ApiService is a service that implements the logic for the ApacheZookeeperClusterV2ApiServicer
// This service should implement the business logic for every endpoint for the ApacheZookeeperClusterV2Api API.
// Include any external packages or services that will be required by this service.
type ApacheZookeeperClusterV2ApiService struct {
}

// NewApacheZookeeperClusterV2ApiService creates a default api service
func NewApacheZookeeperClusterV2ApiService() ApacheZookeeperClusterV2ApiServicer {
	return &ApacheZookeeperClusterV2ApiService{}
}

// ClusterManagementV2ResourcesApplicationsZookeeperClustersV2ClusterIdDelete - Delete cluster
func (s *ApacheZookeeperClusterV2ApiService) ClusterManagementV2ResourcesApplicationsZookeeperClustersV2ClusterIdDelete(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsZookeeperClustersV2ClusterIdDelete with the required logic for this service method.
	// Add api_apache_zookeeper_cluster_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(204, {}) or use other options such as http.Ok ...
	//return Response(204, nil),nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesApplicationsZookeeperClustersV2ClusterIdDelete method not implemented")
}

// ClusterManagementV2ResourcesApplicationsZookeeperClustersV2ClusterIdGet - Get Zookeeper cluster details.
func (s *ApacheZookeeperClusterV2ApiService) ClusterManagementV2ResourcesApplicationsZookeeperClustersV2ClusterIdGet(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsZookeeperClustersV2ClusterIdGet with the required logic for this service method.
	// Add api_apache_zookeeper_cluster_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, ApacheZookeeperClusterV2{}) or use other options such as http.Ok ...
	//return Response(200, ApacheZookeeperClusterV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesApplicationsZookeeperClustersV2ClusterIdGet method not implemented")
}

// ClusterManagementV2ResourcesApplicationsZookeeperClustersV2Post - Create a Zookeeper cluster.
func (s *ApacheZookeeperClusterV2ApiService) ClusterManagementV2ResourcesApplicationsZookeeperClustersV2Post(ctx context.Context, body ApacheZookeeperClusterV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsZookeeperClustersV2Post with the required logic for this service method.
	// Add api_apache_zookeeper_cluster_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(202, ApacheZookeeperClusterV2{}) or use other options such as http.Ok ...
	//return Response(202, ApacheZookeeperClusterV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesApplicationsZookeeperClustersV2Post method not implemented")
}
