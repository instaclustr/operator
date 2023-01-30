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

// RedisClusterV2ApiController binds http requests to an api service and writes the service results to the http response
type RedisClusterV2ApiController struct {
	service      RedisClusterV2ApiServicer
	errorHandler ErrorHandler
}

// RedisClusterV2ApiOption for how the controller is set up.
type RedisClusterV2ApiOption func(*RedisClusterV2ApiController)

// WithRedisClusterV2ApiErrorHandler inject ErrorHandler into controller
func WithRedisClusterV2ApiErrorHandler(h ErrorHandler) RedisClusterV2ApiOption {
	return func(c *RedisClusterV2ApiController) {
		c.errorHandler = h
	}
}

// NewRedisClusterV2ApiController creates a default api controller
func NewRedisClusterV2ApiController(s RedisClusterV2ApiServicer, opts ...RedisClusterV2ApiOption) Router {
	controller := &RedisClusterV2ApiController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the RedisClusterV2ApiController
func (c *RedisClusterV2ApiController) Routes() Routes {
	return Routes{
		{
			"ClusterManagementV2ResourcesApplicationsRedisClustersV2ClusterIdDelete",
			strings.ToUpper("Delete"),
			"/cluster-management/v2/resources/applications/redis/clusters/v2/{clusterId}",
			c.ClusterManagementV2ResourcesApplicationsRedisClustersV2ClusterIdDelete,
		},
		{
			"ClusterManagementV2ResourcesApplicationsRedisClustersV2ClusterIdGet",
			strings.ToUpper("Get"),
			"/cluster-management/v2/resources/applications/redis/clusters/v2/{clusterId}",
			c.ClusterManagementV2ResourcesApplicationsRedisClustersV2ClusterIdGet,
		},
		{
			"ClusterManagementV2ResourcesApplicationsRedisClustersV2Post",
			strings.ToUpper("Post"),
			"/cluster-management/v2/resources/applications/redis/clusters/v2",
			c.ClusterManagementV2ResourcesApplicationsRedisClustersV2Post,
		},
	}
}

// ClusterManagementV2ResourcesApplicationsRedisClustersV2ClusterIdDelete - Delete cluster
func (c *RedisClusterV2ApiController) ClusterManagementV2ResourcesApplicationsRedisClustersV2ClusterIdDelete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]

	result, err := c.service.ClusterManagementV2ResourcesApplicationsRedisClustersV2ClusterIdDelete(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ClusterManagementV2ResourcesApplicationsRedisClustersV2ClusterIdGet - Get Redis cluster details.
func (c *RedisClusterV2ApiController) ClusterManagementV2ResourcesApplicationsRedisClustersV2ClusterIdGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]

	result, err := c.service.ClusterManagementV2ResourcesApplicationsRedisClustersV2ClusterIdGet(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ClusterManagementV2ResourcesApplicationsRedisClustersV2Post - Create a Redis Cluster
func (c *RedisClusterV2ApiController) ClusterManagementV2ResourcesApplicationsRedisClustersV2Post(w http.ResponseWriter, r *http.Request) {
	bodyParam := RedisClusterV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&bodyParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertRedisClusterV2Required(bodyParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2ResourcesApplicationsRedisClustersV2Post(r.Context(), bodyParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}
