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

// KafkaConnectMirrorV2ApiService is a service that implements the logic for the KafkaConnectMirrorV2ApiServicer
// This service should implement the business logic for every endpoint for the KafkaConnectMirrorV2Api API.
// Include any external packages or services that will be required by this service.
type KafkaConnectMirrorV2ApiService struct {
}

// NewKafkaConnectMirrorV2ApiService creates a default api service
func NewKafkaConnectMirrorV2ApiService() KafkaConnectMirrorV2ApiServicer {
	return &KafkaConnectMirrorV2ApiService{}
}

// ClusterManagementV2DataSourcesKafkaConnectClusterClusterIdMirrorsV2Get - List all Kafka connect mirrors.
func (s *KafkaConnectMirrorV2ApiService) ClusterManagementV2DataSourcesKafkaConnectClusterClusterIdMirrorsV2Get(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2DataSourcesKafkaConnectClusterClusterIdMirrorsV2Get with the required logic for this service method.
	// Add api_kafka_connect_mirror_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, []KafkaConnectMirrorSummariesV2{}) or use other options such as http.Ok ...
	//return Response(200, []KafkaConnectMirrorSummariesV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2DataSourcesKafkaConnectClusterClusterIdMirrorsV2Get method not implemented")
}

// ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdDelete - Delete a Kafka Connect Mirror
func (s *KafkaConnectMirrorV2ApiService) ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdDelete(ctx context.Context, mirrorId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdDelete with the required logic for this service method.
	// Add api_kafka_connect_mirror_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(204, {}) or use other options such as http.Ok ...
	//return Response(204, nil),nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdDelete method not implemented")
}

// ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdGet - Get the details of a kafka connect mirror
func (s *KafkaConnectMirrorV2ApiService) ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdGet(ctx context.Context, mirrorId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdGet with the required logic for this service method.
	// Add api_kafka_connect_mirror_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, KafkaConnectMirrorV2{}) or use other options such as http.Ok ...
	//return Response(200, KafkaConnectMirrorV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdGet method not implemented")
}

// ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdPut - Update a Kafka Connect mirror.
func (s *KafkaConnectMirrorV2ApiService) ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdPut(ctx context.Context, mirrorId string, body KafkaConnectMirrorUpdateV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdPut with the required logic for this service method.
	// Add api_kafka_connect_mirror_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, KafkaConnectMirrorV2{}) or use other options such as http.Ok ...
	//return Response(200, KafkaConnectMirrorV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdPut method not implemented")
}

// ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2Post - Create a Kafka Connect mirror
func (s *KafkaConnectMirrorV2ApiService) ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2Post(ctx context.Context, body KafkaConnectMirrorV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2Post with the required logic for this service method.
	// Add api_kafka_connect_mirror_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(202, KafkaConnectMirrorV2{}) or use other options such as http.Ok ...
	//return Response(202, KafkaConnectMirrorV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2Post method not implemented")
}