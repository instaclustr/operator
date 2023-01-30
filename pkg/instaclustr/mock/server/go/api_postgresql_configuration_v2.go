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

// PostgresqlConfigurationV2ApiController binds http requests to an api service and writes the service results to the http response
type PostgresqlConfigurationV2ApiController struct {
	service      PostgresqlConfigurationV2ApiServicer
	errorHandler ErrorHandler
}

// PostgresqlConfigurationV2ApiOption for how the controller is set up.
type PostgresqlConfigurationV2ApiOption func(*PostgresqlConfigurationV2ApiController)

// WithPostgresqlConfigurationV2ApiErrorHandler inject ErrorHandler into controller
func WithPostgresqlConfigurationV2ApiErrorHandler(h ErrorHandler) PostgresqlConfigurationV2ApiOption {
	return func(c *PostgresqlConfigurationV2ApiController) {
		c.errorHandler = h
	}
}

// NewPostgresqlConfigurationV2ApiController creates a default api controller
func NewPostgresqlConfigurationV2ApiController(s PostgresqlConfigurationV2ApiServicer, opts ...PostgresqlConfigurationV2ApiOption) Router {
	controller := &PostgresqlConfigurationV2ApiController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the PostgresqlConfigurationV2ApiController
func (c *PostgresqlConfigurationV2ApiController) Routes() Routes {
	return Routes{
		{
			"ClusterManagementV2DataSourcesPostgresqlClusterClusterIdConfigurationsGet",
			strings.ToUpper("Get"),
			"/cluster-management/v2/data-sources/postgresql_cluster/{clusterId}/configurations",
			c.ClusterManagementV2DataSourcesPostgresqlClusterClusterIdConfigurationsGet,
		},
		{
			"ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2ConfigurationIdDelete",
			strings.ToUpper("Delete"),
			"/cluster-management/v2/resources/applications/postgresql/configurations/v2/{configurationId}",
			c.ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2ConfigurationIdDelete,
		},
		{
			"ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2ConfigurationIdGet",
			strings.ToUpper("Get"),
			"/cluster-management/v2/resources/applications/postgresql/configurations/v2/{configurationId}",
			c.ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2ConfigurationIdGet,
		},
		{
			"ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2ConfigurationIdPut",
			strings.ToUpper("Put"),
			"/cluster-management/v2/resources/applications/postgresql/configurations/v2/{configurationId}",
			c.ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2ConfigurationIdPut,
		},
		{
			"ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2Post",
			strings.ToUpper("Post"),
			"/cluster-management/v2/resources/applications/postgresql/configurations/v2/",
			c.ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2Post,
		},
	}
}

// ClusterManagementV2DataSourcesPostgresqlClusterClusterIdConfigurationsGet - Get cluster configurations
func (c *PostgresqlConfigurationV2ApiController) ClusterManagementV2DataSourcesPostgresqlClusterClusterIdConfigurationsGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]

	result, err := c.service.ClusterManagementV2DataSourcesPostgresqlClusterClusterIdConfigurationsGet(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2ConfigurationIdDelete - Reset a configuration
func (c *PostgresqlConfigurationV2ApiController) ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2ConfigurationIdDelete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	configurationIdParam := params["configurationId"]

	result, err := c.service.ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2ConfigurationIdDelete(r.Context(), configurationIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2ConfigurationIdGet - Get configuration
func (c *PostgresqlConfigurationV2ApiController) ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2ConfigurationIdGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	configurationIdParam := params["configurationId"]

	result, err := c.service.ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2ConfigurationIdGet(r.Context(), configurationIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2ConfigurationIdPut - Update configuration
func (c *PostgresqlConfigurationV2ApiController) ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2ConfigurationIdPut(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	configurationIdParam := params["configurationId"]

	bodyParam := PostgresqlConfigurationPropertyUpdateV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&bodyParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertPostgresqlConfigurationPropertyUpdateV2Required(bodyParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2ConfigurationIdPut(r.Context(), configurationIdParam, bodyParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2Post - Create configuration
func (c *PostgresqlConfigurationV2ApiController) ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2Post(w http.ResponseWriter, r *http.Request) {
	bodyParam := PostgresqlConfigurationPropertyV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&bodyParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertPostgresqlConfigurationPropertyV2Required(bodyParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2ResourcesApplicationsPostgresqlConfigurationsV2Post(r.Context(), bodyParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}
