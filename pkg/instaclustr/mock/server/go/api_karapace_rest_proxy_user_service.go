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

// KarapaceRestProxyUserAPIService is a service that implements the logic for the KarapaceRestProxyUserAPIServicer
// This service should implement the business logic for every endpoint for the KarapaceRestProxyUserAPI API.
// Include any external packages or services that will be required by this service.
type KarapaceRestProxyUserAPIService struct {
}

// NewKarapaceRestProxyUserAPIService creates a default api service
func NewKarapaceRestProxyUserAPIService() KarapaceRestProxyUserAPIServicer {
	return &KarapaceRestProxyUserAPIService{}
}

// ClusterManagementV2DataSourcesKarapaceRestProxyClusterClusterIdKarapaceRestProxyUsersV2Get - List all Karapace Rest Proxy users.
func (s *KarapaceRestProxyUserAPIService) ClusterManagementV2DataSourcesKarapaceRestProxyClusterClusterIdKarapaceRestProxyUsersV2Get(ctx context.Context, clusterId string) (ImplResponse, error) {
	// TODO - update ClusterManagementV2DataSourcesKarapaceRestProxyClusterClusterIdKarapaceRestProxyUsersV2Get with the required logic for this service method.
	// Add api_karapace_rest_proxy_user_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(200, []KarapaceRestProxyUserSummariesV2{}) or use other options such as http.Ok ...
	// return Response(200, []KarapaceRestProxyUserSummariesV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2DataSourcesKarapaceRestProxyClusterClusterIdKarapaceRestProxyUsersV2Get method not implemented")
}

// ClusterManagementV2OperationsApplicationsKarapaceRestProxyClustersV2ClusterIdUsersV2UserNameChangePasswordV2Put - Change a Karapace Rest Proxy user password.
func (s *KarapaceRestProxyUserAPIService) ClusterManagementV2OperationsApplicationsKarapaceRestProxyClustersV2ClusterIdUsersV2UserNameChangePasswordV2Put(ctx context.Context, clusterId string, userName string, karapaceRestProxyUserPasswordV2 KarapaceRestProxyUserPasswordV2) (ImplResponse, error) {
	// TODO - update ClusterManagementV2OperationsApplicationsKarapaceRestProxyClustersV2ClusterIdUsersV2UserNameChangePasswordV2Put with the required logic for this service method.
	// Add api_karapace_rest_proxy_user_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(202, KarapaceRestProxyUserPasswordV2{}) or use other options such as http.Ok ...
	// return Response(202, KarapaceRestProxyUserPasswordV2{}), nil

	// TODO: Uncomment the next line to return response Response(404, ErrorResponseV2{}) or use other options such as http.Ok ...
	// return Response(404, ErrorResponseV2{}), nil

	// TODO: Uncomment the next line to return response Response(503, ErrorResponseV2{}) or use other options such as http.Ok ...
	// return Response(503, ErrorResponseV2{}), nil

	// TODO: Uncomment the next line to return response Response(504, ErrorResponseV2{}) or use other options such as http.Ok ...
	// return Response(504, ErrorResponseV2{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2OperationsApplicationsKarapaceRestProxyClustersV2ClusterIdUsersV2UserNameChangePasswordV2Put method not implemented")
}