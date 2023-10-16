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

// CadenceProvisioningV2APIController binds http requests to an api service and writes the service results to the http response
type CadenceProvisioningV2APIController struct {
	service      CadenceProvisioningV2APIServicer
	errorHandler ErrorHandler
}

// CadenceProvisioningV2APIOption for how the controller is set up.
type CadenceProvisioningV2APIOption func(*CadenceProvisioningV2APIController)

// WithCadenceProvisioningV2APIErrorHandler inject ErrorHandler into controller
func WithCadenceProvisioningV2APIErrorHandler(h ErrorHandler) CadenceProvisioningV2APIOption {
	return func(c *CadenceProvisioningV2APIController) {
		c.errorHandler = h
	}
}

// NewCadenceProvisioningV2APIController creates a default api controller
func NewCadenceProvisioningV2APIController(s CadenceProvisioningV2APIServicer, opts ...CadenceProvisioningV2APIOption) Router {
	controller := &CadenceProvisioningV2APIController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the CadenceProvisioningV2APIController
func (c *CadenceProvisioningV2APIController) Routes() Routes {
	return Routes{
		"ClusterManagementV2ResourcesApplicationsCadenceClustersV2ClusterIdDelete": Route{
			strings.ToUpper("Delete"),
			"/cluster-management/v2/resources/applications/cadence/clusters/v2/{clusterId}",
			c.ClusterManagementV2ResourcesApplicationsCadenceClustersV2ClusterIdDelete,
		},
		"ClusterManagementV2ResourcesApplicationsCadenceClustersV2ClusterIdGet": Route{
			strings.ToUpper("Get"),
			"/cluster-management/v2/resources/applications/cadence/clusters/v2/{clusterId}",
			c.ClusterManagementV2ResourcesApplicationsCadenceClustersV2ClusterIdGet,
		},
		"ClusterManagementV2ResourcesApplicationsCadenceClustersV2ClusterIdPut": Route{
			strings.ToUpper("Put"),
			"/cluster-management/v2/resources/applications/cadence/clusters/v2/{clusterId}",
			c.ClusterManagementV2ResourcesApplicationsCadenceClustersV2ClusterIdPut,
		},
		"ClusterManagementV2ResourcesApplicationsCadenceClustersV2Post": Route{
			strings.ToUpper("Post"),
			"/cluster-management/v2/resources/applications/cadence/clusters/v2/",
			c.ClusterManagementV2ResourcesApplicationsCadenceClustersV2Post,
		},
		"ClusterManagementV2ResourcesApplicationsVersions": Route{
			strings.ToUpper("Get"),
			"/cluster-management/v2/data-sources/applications/{appKind}/versions/v2/",
			c.ClusterManagementV2ResourcesApplicationsVersions,
		},
	}
}

// ClusterManagementV2ResourcesApplicationsCadenceClustersV2ClusterIdDelete - Delete cluster
func (c *CadenceProvisioningV2APIController) ClusterManagementV2ResourcesApplicationsCadenceClustersV2ClusterIdDelete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]
	result, err := c.service.ClusterManagementV2ResourcesApplicationsCadenceClustersV2ClusterIdDelete(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// ClusterManagementV2ResourcesApplicationsCadenceClustersV2ClusterIdGet - Get Cadence cluster details.
func (c *CadenceProvisioningV2APIController) ClusterManagementV2ResourcesApplicationsCadenceClustersV2ClusterIdGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]
	result, err := c.service.ClusterManagementV2ResourcesApplicationsCadenceClustersV2ClusterIdGet(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// ClusterManagementV2ResourcesApplicationsCadenceClustersV2ClusterIdPut - Update a Cadence cluster
func (c *CadenceProvisioningV2APIController) ClusterManagementV2ResourcesApplicationsCadenceClustersV2ClusterIdPut(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]
	cadenceClusterUpdateV2Param := CadenceClusterUpdateV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&cadenceClusterUpdateV2Param); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertCadenceClusterUpdateV2Required(cadenceClusterUpdateV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	if err := AssertCadenceClusterUpdateV2Constraints(cadenceClusterUpdateV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2ResourcesApplicationsCadenceClustersV2ClusterIdPut(r.Context(), clusterIdParam, cadenceClusterUpdateV2Param)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

// ClusterManagementV2ResourcesApplicationsCadenceClustersV2Post - Create a Cadence cluster
func (c *CadenceProvisioningV2APIController) ClusterManagementV2ResourcesApplicationsCadenceClustersV2Post(w http.ResponseWriter, r *http.Request) {
	cadenceClusterV2Param := CadenceClusterV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&cadenceClusterV2Param); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertCadenceClusterV2Required(cadenceClusterV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	if err := AssertCadenceClusterV2Constraints(cadenceClusterV2Param); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2ResourcesApplicationsCadenceClustersV2Post(r.Context(), cadenceClusterV2Param)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}

func (c *CadenceProvisioningV2APIController) ClusterManagementV2ResourcesApplicationsVersions(w http.ResponseWriter, r *http.Request) {
	appKind := mux.Vars(r)["appKind"]

	result, err := c.service.ClusterManagementV2ResourcesApplicationsVersions(r.Context(), appKind)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}

	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)
}
