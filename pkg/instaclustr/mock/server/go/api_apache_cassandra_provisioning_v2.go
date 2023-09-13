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

// ApacheCassandraProvisioningV2APIController binds http requests to an api service and writes the service results to the http response
type ApacheCassandraProvisioningV2APIController struct {
	service      ApacheCassandraProvisioningV2APIServicer
	errorHandler ErrorHandler
}

// ApacheCassandraProvisioningV2APIOption for how the controller is set up.
type ApacheCassandraProvisioningV2APIOption func(*ApacheCassandraProvisioningV2APIController)

// WithApacheCassandraProvisioningV2APIErrorHandler inject ErrorHandler into controller
func WithApacheCassandraProvisioningV2APIErrorHandler(h ErrorHandler) ApacheCassandraProvisioningV2APIOption {
	return func(c *ApacheCassandraProvisioningV2APIController) {
		c.errorHandler = h
	}
}

// NewApacheCassandraProvisioningV2APIController creates a default api controller
func NewApacheCassandraProvisioningV2APIController(s ApacheCassandraProvisioningV2APIServicer, opts ...ApacheCassandraProvisioningV2APIOption) Router {
	controller := &ApacheCassandraProvisioningV2APIController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the ApacheCassandraProvisioningV2APIController
func (c *ApacheCassandraProvisioningV2APIController) Routes() Routes {
	return Routes{
		"ClusterManagementV2DataSourcesApplicationsCassandraClustersV2ClusterIdListBackupsV2Get": Route{
			strings.ToUpper("Get"),
			"/cluster-management/v2/data-sources/applications/cassandra/clusters/v2/{clusterId}/list-backups/v2",
			c.ClusterManagementV2DataSourcesApplicationsCassandraClustersV2ClusterIdListBackupsV2Get,
		},
		"ClusterManagementV2OperationsApplicationsCassandraClustersV2ClusterIdTriggerBackupV2Post": Route{
			strings.ToUpper("Post"),
			"/cluster-management/v2/operations/applications/cassandra/clusters/v2/{clusterId}/trigger-backup/v2",
			c.ClusterManagementV2OperationsApplicationsCassandraClustersV2ClusterIdTriggerBackupV2Post,
		},
		"ClusterManagementV2OperationsApplicationsCassandraRestoreV2Post": Route{
			strings.ToUpper("Post"),
			"/cluster-management/v2/operations/applications/cassandra/restore/v2",
			c.ClusterManagementV2OperationsApplicationsCassandraRestoreV2Post,
		},
		"ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdDelete": Route{
			strings.ToUpper("Delete"),
			"/cluster-management/v2/resources/applications/cassandra/clusters/v2/{clusterId}",
			c.ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdDelete,
		},
		"ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdGet": Route{
			strings.ToUpper("Get"),
			"/cluster-management/v2/resources/applications/cassandra/clusters/v2/{clusterId}",
			c.ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdGet,
		},
		"ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdPut": Route{
			strings.ToUpper("Put"),
			"/cluster-management/v2/resources/applications/cassandra/clusters/v2/{clusterId}",
			c.ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdPut,
		},
		"ClusterManagementV2ResourcesApplicationsCassandraClustersV2Post": Route{
			strings.ToUpper("Post"),
			"/cluster-management/v2/resources/applications/cassandra/clusters/v2",
			c.ClusterManagementV2ResourcesApplicationsCassandraClustersV2Post,
		},
	}
}

// ClusterManagementV2DataSourcesApplicationsCassandraClustersV2ClusterIdListBackupsV2Get - List recent cluster backup events.
func (c *ApacheCassandraProvisioningV2APIController) ClusterManagementV2DataSourcesApplicationsCassandraClustersV2ClusterIdListBackupsV2Get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]
	result, err := c.service.ClusterManagementV2DataSourcesApplicationsCassandraClustersV2ClusterIdListBackupsV2Get(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// ClusterManagementV2OperationsApplicationsCassandraClustersV2ClusterIdTriggerBackupV2Post - Manually trigger cluster backup.
func (c *ApacheCassandraProvisioningV2APIController) ClusterManagementV2OperationsApplicationsCassandraClustersV2ClusterIdTriggerBackupV2Post(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]
	result, err := c.service.ClusterManagementV2OperationsApplicationsCassandraClustersV2ClusterIdTriggerBackupV2Post(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// ClusterManagementV2OperationsApplicationsCassandraRestoreV2Post - Trigger a Cassandra Cluster Restore
func (c *ApacheCassandraProvisioningV2APIController) ClusterManagementV2OperationsApplicationsCassandraRestoreV2Post(w http.ResponseWriter, r *http.Request) {
	cassandraClusterRestoreV2Param := CassandraClusterRestoreV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&cassandraClusterRestoreV2Param); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertCassandraClusterRestoreV2Required(cassandraClusterRestoreV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	if err := AssertCassandraClusterRestoreV2Constraints(cassandraClusterRestoreV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2OperationsApplicationsCassandraRestoreV2Post(r.Context(), cassandraClusterRestoreV2Param)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdDelete - Delete cluster
func (c *ApacheCassandraProvisioningV2APIController) ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdDelete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]
	result, err := c.service.ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdDelete(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdGet - Get Cassandra cluster details.
func (c *ApacheCassandraProvisioningV2APIController) ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]
	result, err := c.service.ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdGet(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdPut - Update Cassandra Cluster Details
func (c *ApacheCassandraProvisioningV2APIController) ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdPut(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]
	cassandraClusterUpdateV2Param := CassandraClusterUpdateV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&cassandraClusterUpdateV2Param); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertCassandraClusterUpdateV2Required(cassandraClusterUpdateV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	if err := AssertCassandraClusterUpdateV2Constraints(cassandraClusterUpdateV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2ResourcesApplicationsCassandraClustersV2ClusterIdPut(r.Context(), clusterIdParam, cassandraClusterUpdateV2Param)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// ClusterManagementV2ResourcesApplicationsCassandraClustersV2Post - Create a Cassandra cluster.
func (c *ApacheCassandraProvisioningV2APIController) ClusterManagementV2ResourcesApplicationsCassandraClustersV2Post(w http.ResponseWriter, r *http.Request) {
	cassandraClusterV2Param := CassandraClusterV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&cassandraClusterV2Param); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertCassandraClusterV2Required(cassandraClusterV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	if err := AssertCassandraClusterV2Constraints(cassandraClusterV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2ResourcesApplicationsCassandraClustersV2Post(r.Context(), cassandraClusterV2Param)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}