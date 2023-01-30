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

// MongodbClusterV2ApiController binds http requests to an api service and writes the service results to the http response
type MongodbClusterV2ApiController struct {
	service      MongodbClusterV2ApiServicer
	errorHandler ErrorHandler
}

// MongodbClusterV2ApiOption for how the controller is set up.
type MongodbClusterV2ApiOption func(*MongodbClusterV2ApiController)

// WithMongodbClusterV2ApiErrorHandler inject ErrorHandler into controller
func WithMongodbClusterV2ApiErrorHandler(h ErrorHandler) MongodbClusterV2ApiOption {
	return func(c *MongodbClusterV2ApiController) {
		c.errorHandler = h
	}
}

// NewMongodbClusterV2ApiController creates a default api controller
func NewMongodbClusterV2ApiController(s MongodbClusterV2ApiServicer, opts ...MongodbClusterV2ApiOption) Router {
	controller := &MongodbClusterV2ApiController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the MongodbClusterV2ApiController
func (c *MongodbClusterV2ApiController) Routes() Routes {
	return Routes{
		{
			"ClusterManagementV2ResourcesApplicationsMongodbClustersV2ClusterIdDelete",
			strings.ToUpper("Delete"),
			"/cluster-management/v2/resources/applications/mongodb/clusters/v2/{clusterId}",
			c.ClusterManagementV2ResourcesApplicationsMongodbClustersV2ClusterIdDelete,
		},
		{
			"ClusterManagementV2ResourcesApplicationsMongodbClustersV2ClusterIdGet",
			strings.ToUpper("Get"),
			"/cluster-management/v2/resources/applications/mongodb/clusters/v2/{clusterId}",
			c.ClusterManagementV2ResourcesApplicationsMongodbClustersV2ClusterIdGet,
		},
		{
			"ClusterManagementV2ResourcesApplicationsMongodbClustersV2Post",
			strings.ToUpper("Post"),
			"/cluster-management/v2/resources/applications/mongodb/clusters/v2",
			c.ClusterManagementV2ResourcesApplicationsMongodbClustersV2Post,
		},
	}
}

// ClusterManagementV2ResourcesApplicationsMongodbClustersV2ClusterIdDelete - Delete cluster
func (c *MongodbClusterV2ApiController) ClusterManagementV2ResourcesApplicationsMongodbClustersV2ClusterIdDelete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]

	result, err := c.service.ClusterManagementV2ResourcesApplicationsMongodbClustersV2ClusterIdDelete(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ClusterManagementV2ResourcesApplicationsMongodbClustersV2ClusterIdGet - Get MongoDB cluster details.
func (c *MongodbClusterV2ApiController) ClusterManagementV2ResourcesApplicationsMongodbClustersV2ClusterIdGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]

	result, err := c.service.ClusterManagementV2ResourcesApplicationsMongodbClustersV2ClusterIdGet(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ClusterManagementV2ResourcesApplicationsMongodbClustersV2Post - Create a MongoDB cluster.
func (c *MongodbClusterV2ApiController) ClusterManagementV2ResourcesApplicationsMongodbClustersV2Post(w http.ResponseWriter, r *http.Request) {
	bodyParam := MongodbClusterV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&bodyParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertMongodbClusterV2Required(bodyParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2ResourcesApplicationsMongodbClustersV2Post(r.Context(), bodyParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}
