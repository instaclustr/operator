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

// KafkaUserV2ApiService is a service that implements the logic for the KafkaUserV2ApiServicer
// This service should implement the business logic for every endpoint for the KafkaUserV2Api API.
// Include any external packages or services that will be required by this service.
type KafkaUserV2ApiService struct {
}

// NewKafkaUserV2ApiService creates a default api service
func NewKafkaUserV2ApiService() KafkaUserV2ApiServicer {
	return &KafkaUserV2ApiService{}
}

// ClusterManagementV2DataSourcesKafkaClusterClusterIdUsersV2Get - List all Kafka users.
func (s *KafkaUserV2ApiService) ClusterManagementV2DataSourcesKafkaClusterClusterIdUsersV2Get(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2DataSourcesKafkaClusterClusterIdUsersV2Get with the required logic for this service method.
	// Add api_kafka_user_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, []KafkaUserSummariesV2{}) or use other options such as http.Ok ...
	//return Response(200, []KafkaUserSummariesV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2DataSourcesKafkaClusterClusterIdUsersV2Get method not implemented")
}

// ClusterManagementV2ResourcesApplicationsKafkaUsersV2KafkaUserIdDelete - Delete a Kafka user
func (s *KafkaUserV2ApiService) ClusterManagementV2ResourcesApplicationsKafkaUsersV2KafkaUserIdDelete(ctx context.Context, kafkaUserId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsKafkaUsersV2KafkaUserIdDelete with the required logic for this service method.
	// Add api_kafka_user_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(204, {}) or use other options such as http.Ok ...
	//return Response(204, nil),nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesApplicationsKafkaUsersV2KafkaUserIdDelete method not implemented")
}

// ClusterManagementV2ResourcesApplicationsKafkaUsersV2KafkaUserIdGet - Get Kafka User details.
func (s *KafkaUserV2ApiService) ClusterManagementV2ResourcesApplicationsKafkaUsersV2KafkaUserIdGet(ctx context.Context, kafkaUserId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsKafkaUsersV2KafkaUserIdGet with the required logic for this service method.
	// Add api_kafka_user_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, KafkaUserV2{}) or use other options such as http.Ok ...
	//return Response(200, KafkaUserV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesApplicationsKafkaUsersV2KafkaUserIdGet method not implemented")
}

// ClusterManagementV2ResourcesApplicationsKafkaUsersV2KafkaUserIdPut - Update Kafka user password
func (s *KafkaUserV2ApiService) ClusterManagementV2ResourcesApplicationsKafkaUsersV2KafkaUserIdPut(ctx context.Context, kafkaUserId string, body KafkaUserV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsKafkaUsersV2KafkaUserIdPut with the required logic for this service method.
	// Add api_kafka_user_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, KafkaUserV2{}) or use other options such as http.Ok ...
	//return Response(200, KafkaUserV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesApplicationsKafkaUsersV2KafkaUserIdPut method not implemented")
}

// ClusterManagementV2ResourcesApplicationsKafkaUsersV2Post - Create a Kafka User.
func (s *KafkaUserV2ApiService) ClusterManagementV2ResourcesApplicationsKafkaUsersV2Post(ctx context.Context, body KafkaUserV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsKafkaUsersV2Post with the required logic for this service method.
	// Add api_kafka_user_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(202, KafkaUserV2{}) or use other options such as http.Ok ...
	//return Response(202, KafkaUserV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesApplicationsKafkaUsersV2Post method not implemented")
}