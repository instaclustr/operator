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

// KafkaConnectMirrorV2ApiController binds http requests to an api service and writes the service results to the http response
type KafkaConnectMirrorV2ApiController struct {
	service      KafkaConnectMirrorV2ApiServicer
	errorHandler ErrorHandler
}

// KafkaConnectMirrorV2ApiOption for how the controller is set up.
type KafkaConnectMirrorV2ApiOption func(*KafkaConnectMirrorV2ApiController)

// WithKafkaConnectMirrorV2ApiErrorHandler inject ErrorHandler into controller
func WithKafkaConnectMirrorV2ApiErrorHandler(h ErrorHandler) KafkaConnectMirrorV2ApiOption {
	return func(c *KafkaConnectMirrorV2ApiController) {
		c.errorHandler = h
	}
}

// NewKafkaConnectMirrorV2ApiController creates a default api controller
func NewKafkaConnectMirrorV2ApiController(s KafkaConnectMirrorV2ApiServicer, opts ...KafkaConnectMirrorV2ApiOption) Router {
	controller := &KafkaConnectMirrorV2ApiController{
		service:      s,
		errorHandler: DefaultErrorHandler,
	}

	for _, opt := range opts {
		opt(controller)
	}

	return controller
}

// Routes returns all the api routes for the KafkaConnectMirrorV2ApiController
func (c *KafkaConnectMirrorV2ApiController) Routes() Routes {
	return Routes{
		{
			"ClusterManagementV2DataSourcesKafkaConnectClusterClusterIdMirrorsV2Get",
			strings.ToUpper("Get"),
			"/cluster-management/v2/data-sources/kafka_connect_cluster/{clusterId}/mirrors/v2/",
			c.ClusterManagementV2DataSourcesKafkaConnectClusterClusterIdMirrorsV2Get,
		},
		{
			"ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdDelete",
			strings.ToUpper("Delete"),
			"/cluster-management/v2/resources/applications/kafka_connect/mirrors/v2/{mirrorId}",
			c.ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdDelete,
		},
		{
			"ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdGet",
			strings.ToUpper("Get"),
			"/cluster-management/v2/resources/applications/kafka_connect/mirrors/v2/{mirrorId}",
			c.ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdGet,
		},
		{
			"ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdPut",
			strings.ToUpper("Put"),
			"/cluster-management/v2/resources/applications/kafka_connect/mirrors/v2/{mirrorId}",
			c.ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdPut,
		},
		{
			"ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2Post",
			strings.ToUpper("Post"),
			"/cluster-management/v2/resources/applications/kafka_connect/mirrors/v2",
			c.ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2Post,
		},
	}
}

// ClusterManagementV2DataSourcesKafkaConnectClusterClusterIdMirrorsV2Get - List all Kafka connect mirrors.
func (c *KafkaConnectMirrorV2ApiController) ClusterManagementV2DataSourcesKafkaConnectClusterClusterIdMirrorsV2Get(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	clusterIdParam := params["clusterId"]

	result, err := c.service.ClusterManagementV2DataSourcesKafkaConnectClusterClusterIdMirrorsV2Get(r.Context(), clusterIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdDelete - Delete a Kafka Connect Mirror
func (c *KafkaConnectMirrorV2ApiController) ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdDelete(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	mirrorIdParam := params["mirrorId"]

	result, err := c.service.ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdDelete(r.Context(), mirrorIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdGet - Get the details of a kafka connect mirror
func (c *KafkaConnectMirrorV2ApiController) ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	mirrorIdParam := params["mirrorId"]

	result, err := c.service.ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdGet(r.Context(), mirrorIdParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdPut - Update a Kafka Connect mirror.
func (c *KafkaConnectMirrorV2ApiController) ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdPut(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	mirrorIdParam := params["mirrorId"]

	bodyParam := KafkaConnectMirrorUpdateV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&bodyParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertKafkaConnectMirrorUpdateV2Required(bodyParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2MirrorIdPut(r.Context(), mirrorIdParam, bodyParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}

// ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2Post - Create a Kafka Connect mirror
func (c *KafkaConnectMirrorV2ApiController) ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2Post(w http.ResponseWriter, r *http.Request) {
	bodyParam := KafkaConnectMirrorV2{}
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&bodyParam); err != nil {
		c.errorHandler(w, r, &ParsingError{Err: err}, nil)
		return
	}
	if err := AssertKafkaConnectMirrorV2Required(bodyParam); err != nil {
		c.errorHandler(w, r, err, nil)
		return
	}
	result, err := c.service.ClusterManagementV2ResourcesApplicationsKafkaConnectMirrorsV2Post(r.Context(), bodyParam)
	// If an error occurred, encode the error with the status code
	if err != nil {
		c.errorHandler(w, r, err, &result)
		return
	}
	// If no error, encode the body and the result code
	EncodeJSONResponse(result.Body, &result.Code, w)

}