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
	"fmt"
	"net/http"

	"k8s.io/utils/strings/slices"
)

// BundleUserAPIService is a service that implements the logic for the BundleUserAPIServicer
// This service should implement the business logic for every endpoint for the BundleUserAPI API.
// Include any external packages or services that will be required by this service.
type BundleUserAPIService struct {
	users map[string][]string
}

// NewBundleUserAPIService creates a default api service
func NewBundleUserAPIService() BundleUserAPIServicer {
	return &BundleUserAPIService{users: make(map[string][]string)}
}

// CreateUser - Add a bundle user
func (s *BundleUserAPIService) CreateUser(ctx context.Context, clusterId string, bundle string, bundleUserCreateRequest BundleUserCreateRequest) (ImplResponse, error) {
	// TODO - update CreateUser with the required logic for this service method.
	// Add api_bundle_user_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(404, ErrorMessage{}) or use other options such as http.Ok ...
	// return Response(404, ErrorMessage{}), nil

	if slices.Contains(s.users[clusterId], bundleUserCreateRequest.Username) {
		return Response(400, GenericResponse{fmt.Sprintf("The user already exists, Username: %s", bundleUserCreateRequest.Username)}), nil
	}
	s.users[clusterId] = append(s.users[clusterId], bundleUserCreateRequest.Username)

	return Response(201, GenericResponse{}), nil
}

// DeleteUser - Delete a bundle user
func (s *BundleUserAPIService) DeleteUser(ctx context.Context, clusterId string, bundle string, bundleUserDeleteRequest BundleUserDeleteRequest) (ImplResponse, error) {
	// TODO - update DeleteUser with the required logic for this service method.
	// Add api_bundle_user_service.go to the .openapi-generator-ignore to avoid overwriting this service implementation when updating open api generation.

	// TODO: Uncomment the next line to return response Response(400, ErrorMessage{}) or use other options such as http.Ok ...
	// return Response(400, ErrorMessage{}), nil

	// TODO: Uncomment the next line to return response Response(401, {}) or use other options such as http.Ok ...
	// return Response(401, nil),nil

	// TODO: Uncomment the next line to return response Response(403, ErrorMessage{}) or use other options such as http.Ok ...
	// return Response(403, ErrorMessage{}), nil

	// TODO: Uncomment the next line to return response Response(404, ErrorMessage{}) or use other options such as http.Ok ...
	// return Response(404, ErrorMessage{}), nil

	// TODO: Uncomment the next line to return response Response(415, ErrorMessage{}) or use other options such as http.Ok ...
	// return Response(415, ErrorMessage{}), nil

	// TODO: Uncomment the next line to return response Response(429, ErrorMessage{}) or use other options such as http.Ok ...
	// return Response(429, ErrorMessage{}), nil

	// TODO: Uncomment the next line to return response Response(503, {}) or use other options such as http.Ok ...
	// return Response(503, nil),nil

	// TODO: Uncomment the next line to return response Response(504, {}) or use other options such as http.Ok ...
	// return Response(504, nil),nil

	return Response(200, GenericResponse{}), nil
}

// FetchUsers - Fetch a bundle users
func (s *BundleUserAPIService) FetchUsers(ctx context.Context, clusterId string, bundle string) (ImplResponse, error) {

	users := s.users[clusterId]

	return Response(200, users), nil
}

func (s *BundleUserAPIService) GetDefaultCreds(ctx context.Context, clusterID string) (ImplResponse, error) {
	creds := &struct {
		Username                string `json:"username"`
		InstaclustrUserPassword string `json:"instaclustrUserPassword"`
	}{
		Username:                clusterID + "_username",
		InstaclustrUserPassword: clusterID + "_password",
	}

	return Response(http.StatusAccepted, creds), nil
}
