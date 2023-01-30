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
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// ApacheZookeeperClusterV2ApiController binds http requests to an api service and writes the service results to the http response
type ApacheZookeeperClusterV2ApiController struct {
	service      ApacheZookeeperClusterV2ApiServicer
	errorHandler ErrorHandler
}

// ApacheZookeeperClusterV2ApiOption for how the controller is set up.
type ApacheZookeeperClusterV2ApiOption func(*ApacheZookeeperClusterV2ApiController)

// WithApacheZookeeperClusterV2ApiErrorHandler inject ErrorHandler into controller
func WithApacheZookeeperClusterV2ApiErrorHandler(h ErrorHandler) ApacheZookeeperClusterV2ApiOption {
	return func(c *ApacheZookeeperClusterV2ApiController) {
		c.errorHandler = h
	}
}

// NewApacheZookeeperClusterV2ApiController creates a default api controller
func NewApacheZookeeperClusterV2ApiController(s ApacheZookeeperClusterV2ApiServicer, opts ...ApacheZookeeperClusterV2ApiOption) Router {
	controller := &ApacheZookeeperClusterV2ApiController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the ApacheZookeeperClusterV2ApiController
func (c *ApacheZookeeperClusterV2ApiController) Routes() Routes {
	return Routes{
		{
			"ClusterManagementV2ResourcesApplicationsZookeeperClustersV2ClusterIdDelete",
			strings.ToUpper("Delete"),
			"/cluster-management/v2/resources/applications/zookeeper/clusters/v2/{clusterId}",
			c.ClusterManagementV2ResourcesApplicationsZookeeperClustersV2ClusterIdDelete,
		},
		{
			"ClusterManagementV2ResourcesApplicationsZookeeperClustersV2ClusterIdGet",
			strings.ToUpper("Get"),
			"/cluster-management/v2/resources/applications/zookeeper/clusters/v2/{clusterId}",
			c.ClusterManagementV2ResourcesApplicationsZookeeperClustersV2ClusterIdGet,
		},
		{
			"ClusterManagementV2ResourcesApplicationsZookeeperClustersV2Post",
			strings.ToUpper("Post"),
			"/cluster-management/v2/resources/applications/zookeeper/clusters/v2",
			c.ClusterManagementV2ResourcesApplicationsZookeeperClustersV2Post,
		},
	}
}

// ClusterManagementV2ResourcesApplicationsZookeeperClustersV2ClusterIdDelete - Delete cluster
func (c *ApacheZookeeperClusterV2ApiController) ClusterManagementV2ResourcesApplicationsZookeeperClustersV2ClusterIdDelete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]

	result, err := c.service.ClusterManagementV2ResourcesApplicationsZookeeperClustersV2ClusterIdDelete(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ClusterManagementV2ResourcesApplicationsZookeeperClustersV2ClusterIdGet - Get Zookeeper cluster details.
func (c *ApacheZookeeperClusterV2ApiController) ClusterManagementV2ResourcesApplicationsZookeeperClustersV2ClusterIdGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]

	result, err := c.service.ClusterManagementV2ResourcesApplicationsZookeeperClustersV2ClusterIdGet(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ClusterManagementV2ResourcesApplicationsZookeeperClustersV2Post - Create a Zookeeper cluster.
func (c *ApacheZookeeperClusterV2ApiController) ClusterManagementV2ResourcesApplicationsZookeeperClustersV2Post(w http.ResponseWriter, r *http.Request) {
	bodyParam := ApacheZookeeperClusterV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&bodyParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertApacheZookeeperClusterV2Required(bodyParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2ResourcesApplicationsZookeeperClustersV2Post(r.Context(), bodyParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}
