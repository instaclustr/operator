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

// OpenAPISpecificationV2ApiService is a service that implements the logic for the OpenAPISpecificationV2ApiServicer
// This service should implement the business logic for every endpoint for the OpenAPISpecificationV2Api API.
// Include any external packages or services that will be required by this service.
type OpenAPISpecificationV2ApiService struct {
}

// NewOpenAPISpecificationV2ApiService creates a default api service
func NewOpenAPISpecificationV2ApiService() OpenAPISpecificationV2ApiServicer {
	return &OpenAPISpecificationV2ApiService{}
}

// ClusterManagementV2OpenApiSpecificationCurrentGet - Get current endpoints
func (s *OpenAPISpecificationV2ApiService) ClusterManagementV2OpenApiSpecificationCurrentGet(ctx context.Context) (ImplResponse, error) {
	// TODO - update ClusterManagementV2OpenApiSpecificationCurrentGet with the required logic for this service method.
	// Add api_open_api_specification_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, string{}) or use other options such as http.Ok ...
	//return Response(200, string{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2OpenApiSpecificationCurrentGet method not implemented")
}

// ClusterManagementV2OpenApiSpecificationDeprecatedGet - Get deprecated endpoints
func (s *OpenAPISpecificationV2ApiService) ClusterManagementV2OpenApiSpecificationDeprecatedGet(ctx context.Context) (ImplResponse, error) {
	// TODO - update ClusterManagementV2OpenApiSpecificationDeprecatedGet with the required logic for this service method.
	// Add api_open_api_specification_v2_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	//TODO: Uncomment the next line to return response Response(200, string{}) or use other options such as http.Ok ...
	//return Response(200, string{}), nil

	return Response(http.StatusNotImplemented, nil), errors.New("ClusterManagementV2OpenApiSpecificationDeprecatedGet method not implemented")
}