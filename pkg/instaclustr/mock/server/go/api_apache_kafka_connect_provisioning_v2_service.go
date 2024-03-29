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
)

// ApacheKafkaConnectProvisioningV2APIService is a service that implements the logic for the ApacheKafkaConnectProvisioningV2APIServicer
// This service should implement the business logic for every endpoint for the ApacheKafkaConnectProvisioningV2API API.
// Include any external packages or services that will be required by this service.
type ApacheKafkaConnectProvisioningV2APIService struct {
	clusters map[string]*KafkaConnectClusterV2
}

// NewApacheKafkaConnectProvisioningV2APIService creates a default api service
func NewApacheKafkaConnectProvisioningV2APIService() ApacheKafkaConnectProvisioningV2APIServicer {
	return &ApacheKafkaConnectProvisioningV2APIService{clusters: map[string]*KafkaConnectClusterV2{}}
}

// ClusterManagementV2OperationsApplicationsKafkaConnectClustersV2ClusterIdSyncCustomKafkaConnectorsV2Put - Update the custom connectors of a Kafka Connect cluster.
func (s *ApacheKafkaConnectProvisioningV2APIService) ClusterManagementV2OperationsApplicationsKafkaConnectClustersV2ClusterIdSyncCustomKafkaConnectorsV2Put(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2OperationsApplicationsKafkaConnectClustersV2ClusterIdSyncCustomKafkaConnectorsV2Put with the required logic for this service method.
	// Add api_apache_kafka_connect_provisioning_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(202, {}) or use other options such as http.Ok ...
	// return Response(202, nil),nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2OperationsApplicationsKafkaConnectClustersV2ClusterIdSyncCustomKafkaConnectorsV2Put method not implemented")
}

// ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2ClusterIdDelete - Delete Kafka connect Cluster
func (s *ApacheKafkaConnectProvisioningV2APIService) ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2ClusterIdDelete(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2ClusterIdDelete with the required logic for this service method.
	// Add api_apache_kafka_connect_provisioning_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	delete(s.clusters, clusterId)
	return Response(204, nil), nil
}

// ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2ClusterIdGet - Get Kafka connect cluster details
func (s *ApacheKafkaConnectProvisioningV2APIService) ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2ClusterIdGet(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2ClusterIdGet with the required logic for this service method.
	// Add api_apache_kafka_connect_provisioning_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	cluster, exists := s.clusters[clusterId]
	if !exists {
		return Response(404, nil), nil
	}

	cluster.Status = RUNNING

	return Response(200, cluster), nil
}

// ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2ClusterIdPut - Update a Kafka connect cluster
func (s *ApacheKafkaConnectProvisioningV2APIService) ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2ClusterIdPut(ctx context.Context, clusterId string, kafkaConnectClusterUpdateV2 KafkaConnectClusterUpdateV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2ClusterIdPut with the required logic for this service method.
	// Add api_apache_kafka_connect_provisioning_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	cluster, exists := s.clusters[clusterId]
	if !exists {
		return Response(404, nil), nil
	}

	cluster.DataCentres[0].NumberOfNodes = kafkaConnectClusterUpdateV2.DataCentres[0].NumberOfNodes

	return Response(202, nil), nil

	//TODO: Uncomment the next line to return response Response(404, ErrorListResponseV2{}) or use other options such as http.Ok ...
	//return Response(404, ErrorListResponseV2{}), nil
}

// ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2Post - Create a Kafka connect cluster.
func (s *ApacheKafkaConnectProvisioningV2APIService) ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2Post(ctx context.Context, kafkaConnectClusterV2 KafkaConnectClusterV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsKafkaConnectClustersV2Post with the required logic for this service method.
	// Add api_apache_kafka_connect_provisioning_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	kafkaConnectClusterV2.Id = kafkaConnectClusterV2.Name + CreatedID
	s.clusters[kafkaConnectClusterV2.Id] = &kafkaConnectClusterV2

	return Response(202, kafkaConnectClusterV2), nil
}
