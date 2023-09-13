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

// KarapaceRestProxyUserAPIController binds http requests to an api service and writes the service results to the http response
type KarapaceRestProxyUserAPIController struct {
	service      KarapaceRestProxyUserAPIServicer
	errorHandler ErrorHandler
}

// KarapaceRestProxyUserAPIOption for how the controller is set up.
type KarapaceRestProxyUserAPIOption func(*KarapaceRestProxyUserAPIController)

// WithKarapaceRestProxyUserAPIErrorHandler inject ErrorHandler into controller
func WithKarapaceRestProxyUserAPIErrorHandler(h ErrorHandler) KarapaceRestProxyUserAPIOption {
	return func(c *KarapaceRestProxyUserAPIController) {
		c.errorHandler = h
	}
}

// NewKarapaceRestProxyUserAPIController creates a default api controller
func NewKarapaceRestProxyUserAPIController(s KarapaceRestProxyUserAPIServicer, opts ...KarapaceRestProxyUserAPIOption) Router {
	controller := &KarapaceRestProxyUserAPIController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the KarapaceRestProxyUserAPIController
func (c *KarapaceRestProxyUserAPIController) Routes() Routes {
	return Routes{
		"ClusterManagementV2DataSourcesKarapaceRestProxyClusterClusterIdKarapaceRestProxyUsersV2Get": Route{
			strings.ToUpper("Get"),
			"/cluster-management/v2/data-sources/karapace_rest_proxy_cluster/{clusterId}/karapace-rest-proxy-users/v2/",
			c.ClusterManagementV2DataSourcesKarapaceRestProxyClusterClusterIdKarapaceRestProxyUsersV2Get,
		},
		"ClusterManagementV2OperationsApplicationsKarapaceRestProxyClustersV2ClusterIdUsersV2UserNameChangePasswordV2Put": Route{
			strings.ToUpper("Put"),
			"/cluster-management/v2/operations/applications/karapace-rest-proxy/clusters/v2/{clusterId}/users/v2/{userName}/change-password/v2",
			c.ClusterManagementV2OperationsApplicationsKarapaceRestProxyClustersV2ClusterIdUsersV2UserNameChangePasswordV2Put,
		},
	}
}

// ClusterManagementV2DataSourcesKarapaceRestProxyClusterClusterIdKarapaceRestProxyUsersV2Get - List all Karapace Rest Proxy users.
func (c *KarapaceRestProxyUserAPIController) ClusterManagementV2DataSourcesKarapaceRestProxyClusterClusterIdKarapaceRestProxyUsersV2Get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]
	result, err := c.service.ClusterManagementV2DataSourcesKarapaceRestProxyClusterClusterIdKarapaceRestProxyUsersV2Get(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// ClusterManagementV2OperationsApplicationsKarapaceRestProxyClustersV2ClusterIdUsersV2UserNameChangePasswordV2Put - Change a Karapace Rest Proxy user password.
func (c *KarapaceRestProxyUserAPIController) ClusterManagementV2OperationsApplicationsKarapaceRestProxyClustersV2ClusterIdUsersV2UserNameChangePasswordV2Put(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]
	userNameParam := params["userName"]
	karapaceRestProxyUserPasswordV2Param := KarapaceRestProxyUserPasswordV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&karapaceRestProxyUserPasswordV2Param); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertKarapaceRestProxyUserPasswordV2Required(karapaceRestProxyUserPasswordV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	if err := AssertKarapaceRestProxyUserPasswordV2Constraints(karapaceRestProxyUserPasswordV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2OperationsApplicationsKarapaceRestProxyClustersV2ClusterIdUsersV2UserNameChangePasswordV2Put(r.Context(), clusterIdParam, userNameParam, karapaceRestProxyUserPasswordV2Param)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}