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

// PostgresqlClusterV2ApiController binds http requests to an api service and writes the service results to the http response
type PostgresqlClusterV2ApiController struct {
	service      PostgresqlClusterV2ApiServicer
	errorHandler ErrorHandler
}

// PostgresqlClusterV2ApiOption for how the controller is set up.
type PostgresqlClusterV2ApiOption func(*PostgresqlClusterV2ApiController)

// WithPostgresqlClusterV2ApiErrorHandler inject ErrorHandler into controller
func WithPostgresqlClusterV2ApiErrorHandler(h ErrorHandler) PostgresqlClusterV2ApiOption {
	return func(c *PostgresqlClusterV2ApiController) {
		c.errorHandler = h
	}
}

// NewPostgresqlClusterV2ApiController creates a default api controller
func NewPostgresqlClusterV2ApiController(s PostgresqlClusterV2ApiServicer, opts ...PostgresqlClusterV2ApiOption) Router {
	controller := &PostgresqlClusterV2ApiController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the PostgresqlClusterV2ApiController
func (c *PostgresqlClusterV2ApiController) Routes() Routes {
	return Routes{
		{
			"ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2ClusterIdDelete",
			strings.ToUpper("Delete"),
			"/cluster-management/v2/resources/applications/postgresql/clusters/v2/{clusterId}",
			c.ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2ClusterIdDelete,
		},
		{
			"ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2ClusterIdGet",
			strings.ToUpper("Get"),
			"/cluster-management/v2/resources/applications/postgresql/clusters/v2/{clusterId}",
			c.ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2ClusterIdGet,
		},
		{
			"ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2ClusterIdPut",
			strings.ToUpper("Put"),
			"/cluster-management/v2/resources/applications/postgresql/clusters/v2/{clusterId}",
			c.ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2ClusterIdPut,
		},
		{
			"ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2Post",
			strings.ToUpper("Post"),
			"/cluster-management/v2/resources/applications/postgresql/clusters/v2",
			c.ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2Post,
		},
		{
			"ClusterManagementConfigurationV1ResourcesApplicationsPostgresqlClustersV2Put",
			strings.ToUpper("Put"),
			"/provisioning/v1/{clusterId}",
			c.ClusterManagementConfigurationV1ResourcesApplicationsPostgresqlClustersV2Put,
		},
		{
			"ClusterManagementMaintenanceEventsV1ResourcesApplicationsPostgresqlClustersV2Get",
			strings.ToUpper("Get"),
			"/v1/maintenance-events/events",
			c.ClusterManagementMaintenanceEventsV1ResourcesApplicationsPostgresqlClustersV2Get,
		},
	}
}

// ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2ClusterIdDelete - Delete cluster
func (c *PostgresqlClusterV2ApiController) ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2ClusterIdDelete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]

	result, err := c.service.ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2ClusterIdDelete(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2ClusterIdGet - Get Postgresql cluster details.
func (c *PostgresqlClusterV2ApiController) ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2ClusterIdGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]

	result, err := c.service.ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2ClusterIdGet(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2ClusterIdPut - Update PostgreSQL cluster details
func (c *PostgresqlClusterV2ApiController) ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2ClusterIdPut(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]

	bodyParam := PostgresqlClusterUpdateV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&bodyParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertPostgresqlClusterUpdateV2Required(bodyParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2ClusterIdPut(r.Context(), clusterIdParam, bodyParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2Post - Create a Postgresql cluster.
func (c *PostgresqlClusterV2ApiController) ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2Post(w http.ResponseWriter, r *http.Request) {
	bodyParam := PostgresqlClusterV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&bodyParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertPostgresqlClusterV2Required(bodyParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2ResourcesApplicationsPostgresqlClustersV2Post(r.Context(), bodyParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ClusterManagementConfigurationV1ResourcesApplicationsPostgresqlClustersV2Put - Update PostgreSQL cluster details
func (c *PostgresqlClusterV2ApiController) ClusterManagementConfigurationV1ResourcesApplicationsPostgresqlClustersV2Put(w http.ResponseWriter, r *http.Request) {
	// TODO: Remove this method when UpdateDescriptionAndTwoFactorDelete is implemented for API v 2.

	result, err := c.service.ClusterManagementConfigurationV1ResourcesApplicationsPostgresqlClustersV2Put(r.Context())
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ClusterManagementConfigurationV1ResourcesApplicationsPostgresqlClustersV2Put - Update PostgreSQL cluster details
func (c *PostgresqlClusterV2ApiController) ClusterManagementMaintenanceEventsV1ResourcesApplicationsPostgresqlClustersV2Get(w http.ResponseWriter, r *http.Request) {
	// TODO: Remove this method when GetMaintenance is implemented for API v 2.

	clusterIdParam := r.URL.Query().Get("clusterId")

	result, err := c.service.ClusterManagementMaintenanceEventsV1ResourcesApplicationsPostgresqlClustersV2Get(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
	//   'https://api.instaclustr.com/v1/maintenance-events/events?clusterId=string'
}
