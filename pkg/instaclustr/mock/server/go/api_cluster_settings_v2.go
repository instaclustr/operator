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

// ClusterSettingsV2APIController binds http requests to an api service and writes the service results to the http response
type ClusterSettingsV2APIController struct {
	service      ClusterSettingsV2APIServicer
	errorHandler ErrorHandler
}

// ClusterSettingsV2APIOption for how the controller is set up.
type ClusterSettingsV2APIOption func(*ClusterSettingsV2APIController)

// WithClusterSettingsV2APIErrorHandler inject ErrorHandler into controller
func WithClusterSettingsV2APIErrorHandler(h ErrorHandler) ClusterSettingsV2APIOption {
	return func(c *ClusterSettingsV2APIController) {
		c.errorHandler = h
	}
}

// NewClusterSettingsV2APIController creates a default api controller
func NewClusterSettingsV2APIController(s ClusterSettingsV2APIServicer, opts ...ClusterSettingsV2APIOption) Router {
	controller := &ClusterSettingsV2APIController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the ClusterSettingsV2APIController
func (c *ClusterSettingsV2APIController) Routes() Routes {
	return Routes{
		"ClusterManagementV2OperationsClustersV2ClusterIdChangeSettingsV2Put": Route{
			strings.ToUpper("Put"),
			"/cluster-management/v2/operations/clusters/v2/{clusterId}/change-settings/v2",
			c.ClusterManagementV2OperationsClustersV2ClusterIdChangeSettingsV2Put,
		},
	}
}

// ClusterManagementV2OperationsClustersV2ClusterIdChangeSettingsV2Put - Update cluster's general settings
func (c *ClusterSettingsV2APIController) ClusterManagementV2OperationsClustersV2ClusterIdChangeSettingsV2Put(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]
	clusterSettingsUpdateV2Param := ClusterSettingsUpdateV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&clusterSettingsUpdateV2Param); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertClusterSettingsUpdateV2Required(clusterSettingsUpdateV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	if err := AssertClusterSettingsUpdateV2Constraints(clusterSettingsUpdateV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2OperationsClustersV2ClusterIdChangeSettingsV2Put(r.Context(), clusterIdParam, clusterSettingsUpdateV2Param)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}
