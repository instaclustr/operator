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

// KafkaConnectClusterV2ApiService is a service that implements the logic for the KafkaConnectClusterV2ApiServicer
// This service should implement the business logic for every endpoint for the KafkaConnectClusterV2Api API.
// Include any external packages or services that will be required by this service.
type KafkaConnectClusterV2ApiService struct {
	MockKafkaConnectCluster *KafkaConnectClusterV2
}

// NewKafkaConnectClusterV2ApiService creates a default api service
func NewKafkaConnectClusterV2ApiService() KafkaConnectClusterV2ApiServicer {
	return &KafkaConnectClusterV2ApiService{}
}

// ClusterManagementV2OperationsApplicationsKafkaConnectClustersV2ClusterIdSyncCustomKafkaConnectorsV2Put - Update the custom connectors of a Kafka Connect cluster.
func (s *KafkaConnectClusterV2ApiService) ClusterManagementV2OperationsApplicationsKafkaConnectClustersV2ClusterIdSyncCustomKafkaConnectorsV2Put(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2OperationsApplicationsKafkaConnectClustersV2ClusterIdSyncCustomKafkaConnectorsV2Put with the required logic for this service method.
	// Add api_kafka_connect_cluster_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(202, {}) or use other options such as http.Ok ...
	//return Response(202, nil),nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2OperationsApplicationsKafkaConnectClustersV2ClusterIdSyncCustomKafkaConnectorsV2Put method not implemented")
}

// ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2ClusterIdDelete - Delete Kafka connect Cluster
func (s *KafkaConnectClusterV2ApiService) ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2ClusterIdDelete(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2ClusterIdDelete with the required logic for this service method.
	// Add api_kafka_connect_cluster_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	s.MockKafkaConnectCluster = nil
	return Response(204, nil), nil
}

// ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2ClusterIdGet - Get Kafka connect cluster details
func (s *KafkaConnectClusterV2ApiService) ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2ClusterIdGet(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2ClusterIdGet with the required logic for this service method.
	// Add api_kafka_connect_cluster_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	if s.MockKafkaConnectCluster != nil {
		s.MockKafkaConnectCluster.Status = RUNNING
	} else {
		return Response(404, nil), nil
	}

	return Response(200, s.MockKafkaConnectCluster), nil
}

// ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2ClusterIdPut - Update a Kafka connect cluster
func (s *KafkaConnectClusterV2ApiService) ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2ClusterIdPut(ctx context.Context, clusterId string, body KafkaConnectClusterUpdateV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2ClusterIdPut with the required logic for this service method.
	// Add api_kafka_connect_cluster_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	s.MockKafkaConnectCluster.DataCentres[0].NumberOfNodes = body.DataCentres[0].NumberOfNodes

	return Response(202, nil), nil

	//TODO: Uncomment the next line to return response Response(404, ErrorListResponseV2{}) or use other options such as http.Ok ...
	//return Response(404, ErrorListResponseV2{}), nil
}

// ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2Post - Create a Kafka connect cluster.
func (s *KafkaConnectClusterV2ApiService) ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2Post(ctx context.Context, body KafkaConnectClusterV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2Post with the required logic for this service method.
	// Add api_kafka_connect_cluster_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	s.MockKafkaConnectCluster = &body
	s.MockKafkaConnectCluster.Id = CreatedID
	return Response(202, s.MockKafkaConnectCluster), nil
}
