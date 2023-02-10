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

// RedisUserV2ApiService is a service that implements the logic for the RedisUserV2ApiServicer
// This service should implement the business logic for every endpoint for the RedisUserV2Api API.
// Include any external packages or services that will be required by this service.
type RedisUserV2ApiService struct {
}

// NewRedisUserV2ApiService creates a default api service
func NewRedisUserV2ApiService() RedisUserV2ApiServicer {
	return &RedisUserV2ApiService{}
}

// ClusterManagementV2ResourcesApplicationsRedisUsersV2Post - Create a Redis User.
func (s *RedisUserV2ApiService) ClusterManagementV2ResourcesApplicationsRedisUsersV2Post(ctx context.Context, body RedisUserV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsRedisUsersV2Post with the required logic for this service method.
	// Add api_redis_user_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(202, RedisUserV2{}) or use other options such as http.Ok ...
	//return Response(202, RedisUserV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesApplicationsRedisUsersV2Post method not implemented")
}

// ClusterManagementV2ResourcesApplicationsRedisUsersV2RedisUserIdDelete - Delete a Redis user
func (s *RedisUserV2ApiService) ClusterManagementV2ResourcesApplicationsRedisUsersV2RedisUserIdDelete(ctx context.Context, redisUserId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsRedisUsersV2RedisUserIdDelete with the required logic for this service method.
	// Add api_redis_user_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(204, {}) or use other options such as http.Ok ...
	//return Response(204, nil),nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesApplicationsRedisUsersV2RedisUserIdDelete method not implemented")
}

// ClusterManagementV2ResourcesApplicationsRedisUsersV2RedisUserIdGet - Get Redis User details.
func (s *RedisUserV2ApiService) ClusterManagementV2ResourcesApplicationsRedisUsersV2RedisUserIdGet(ctx context.Context, redisUserId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsRedisUsersV2RedisUserIdGet with the required logic for this service method.
	// Add api_redis_user_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, RedisUserV2{}) or use other options such as http.Ok ...
	//return Response(200, RedisUserV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesApplicationsRedisUsersV2RedisUserIdGet method not implemented")
}

// ClusterManagementV2ResourcesApplicationsRedisUsersV2RedisUserIdPut - Update Redis user password
func (s *RedisUserV2ApiService) ClusterManagementV2ResourcesApplicationsRedisUsersV2RedisUserIdPut(ctx context.Context, redisUserId string, body RedisUserPasswordV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2ResourcesApplicationsRedisUsersV2RedisUserIdPut with the required logic for this service method.
	// Add api_redis_user_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, RedisUserPasswordV2{}) or use other options such as http.Ok ...
	//return Response(200, RedisUserPasswordV2{}), nil

	//TODO: Uncomment the next line to return response Response(404, ErrorResponseV2{}) or use other options such as http.Ok ...
	//return Response(404, ErrorResponseV2{}), nil

	//TODO: Uncomment the next line to return response Response(503, ErrorResponseV2{}) or use other options such as http.Ok ...
	//return Response(503, ErrorResponseV2{}), nil

	//TODO: Uncomment the next line to return response Response(504, ErrorResponseV2{}) or use other options such as http.Ok ...
	//return Response(504, ErrorResponseV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2ResourcesApplicationsRedisUsersV2RedisUserIdPut method not implemented")
}