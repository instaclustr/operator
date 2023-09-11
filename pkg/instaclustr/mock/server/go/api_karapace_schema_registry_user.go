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
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// KarapaceSchemaRegistryUserAPIController binds http requests to an api service and writes the service results to the http response
type KarapaceSchemaRegistryUserAPIController struct {
	service      KarapaceSchemaRegistryUserAPIServicer
	errorHandler ErrorHandler
}

// KarapaceSchemaRegistryUserAPIOption for how the controller is set up.
type KarapaceSchemaRegistryUserAPIOption func(*KarapaceSchemaRegistryUserAPIController)

// WithKarapaceSchemaRegistryUserAPIErrorHandler inject ErrorHandler into controller
func WithKarapaceSchemaRegistryUserAPIErrorHandler(h ErrorHandler) KarapaceSchemaRegistryUserAPIOption {
	return func(c *KarapaceSchemaRegistryUserAPIController) {
		c.errorHandler = h
	}
}

// NewKarapaceSchemaRegistryUserAPIController creates a default api controller
func NewKarapaceSchemaRegistryUserAPIController(s KarapaceSchemaRegistryUserAPIServicer, opts ...KarapaceSchemaRegistryUserAPIOption) Router {
	controller := &KarapaceSchemaRegistryUserAPIController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the KarapaceSchemaRegistryUserAPIController
func (c *KarapaceSchemaRegistryUserAPIController) Routes() Routes {
	return Routes{
		"ClusterManagementV2DataSourcesKarapaceSchemaRegistryClusterClusterIdKarapaceSchemaRegistryUsersV2Get": Route{
			strings.ToUpper("Get"),
			"/cluster-management/v2/data-sources/karapace_schema_registry_cluster/{clusterId}/karapace-schema-registry-users/v2/",
			c.ClusterManagementV2DataSourcesKarapaceSchemaRegistryClusterClusterIdKarapaceSchemaRegistryUsersV2Get,
		},
		"ClusterManagementV2OperationsApplicationsKarapaceSchemaRegistryClustersV2ClusterIdUsersV2UserNameChangePasswordV2Put": Route{
			strings.ToUpper("Put"),
			"/cluster-management/v2/operations/applications/karapace-schema-registry/clusters/v2/{clusterId}/users/v2/{userName}/change-password/v2",
			c.ClusterManagementV2OperationsApplicationsKarapaceSchemaRegistryClustersV2ClusterIdUsersV2UserNameChangePasswordV2Put,
		},
	}
}

// ClusterManagementV2DataSourcesKarapaceSchemaRegistryClusterClusterIdKarapaceSchemaRegistryUsersV2Get - List all Karapace Schema Registry users.
func (c *KarapaceSchemaRegistryUserAPIController) ClusterManagementV2DataSourcesKarapaceSchemaRegistryClusterClusterIdKarapaceSchemaRegistryUsersV2Get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]
	result, err := c.service.ClusterManagementV2DataSourcesKarapaceSchemaRegistryClusterClusterIdKarapaceSchemaRegistryUsersV2Get(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// ClusterManagementV2OperationsApplicationsKarapaceSchemaRegistryClustersV2ClusterIdUsersV2UserNameChangePasswordV2Put - Change a Karapace Schema Registry user password.
func (c *KarapaceSchemaRegistryUserAPIController) ClusterManagementV2OperationsApplicationsKarapaceSchemaRegistryClustersV2ClusterIdUsersV2UserNameChangePasswordV2Put(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]
	userNameParam := params["userName"]
	karapaceSchemaRegistryUserPasswordV2Param := KarapaceSchemaRegistryUserPasswordV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&karapaceSchemaRegistryUserPasswordV2Param); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertKarapaceSchemaRegistryUserPasswordV2Required(karapaceSchemaRegistryUserPasswordV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	if err := AssertKarapaceSchemaRegistryUserPasswordV2Constraints(karapaceSchemaRegistryUserPasswordV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2OperationsApplicationsKarapaceSchemaRegistryClustersV2ClusterIdUsersV2UserNameChangePasswordV2Put(r.Context(), clusterIdParam, userNameParam, karapaceSchemaRegistryUserPasswordV2Param)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}
